const API_BASE = '/api';

interface ValidateResponse {
  valid: boolean;
  admin: boolean;
}

interface ApiToken {
  token: string;
  admin: boolean;
}

interface HealthResponse {
  mongo: string;
  redis: string;
}

interface TokenPermissions {
  maintainer: boolean;
  download: boolean;
  upload: boolean;
  delete: boolean;
}

interface Version {
  path: string;
  checksum: string;
}

interface Product {
  name: string;
  tokens: Record<string, TokenPermissions>;
  versions: Record<string, Version>;
}

interface ErrorResponse {
  error: string;
}

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(
  method: string,
  path: string,
  token: string,
  body?: unknown,
  isFormData = false
): Promise<T> {
  const headers: Record<string, string> = {
    'X-Api-Token': token,
  };

  if (!isFormData) {
    headers['Content-Type'] = 'application/json';
  }

  const options: RequestInit = {
    method,
    headers,
  };

  if (body) {
    options.body = isFormData ? body : JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE}${path}`, options);

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({})) as ErrorResponse;
    throw new ApiError(response.status, errorData.error || 'Request failed');
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

export const api = {
  async validateToken(token: string): Promise<ValidateResponse> {
    const response = await fetch(`${API_BASE}/validate`, {
      method: 'GET',
      headers: {
        'X-Api-Token': token,
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({})) as ErrorResponse;
      throw new ApiError(response.status, errorData.error || 'Invalid token');
    }

    return response.json();
  },

  async getHealth(token: string): Promise<HealthResponse> {
    return request<HealthResponse>('GET', '/health', token);
  },

  async registerToken(token: string, admin: boolean): Promise<string> {
    const response = await fetch(`${API_BASE}/register`, {
      method: 'PUT',
      headers: {
        'X-Api-Token': token,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ admin }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({})) as ErrorResponse;
      throw new ApiError(response.status, errorData.error || 'Failed to register token');
    }

    return response.text();
  },

  async fetchToken(token: string, tokenToFetch: string): Promise<ApiToken> {
    return request<ApiToken>('GET', `/fetch/${encodeURIComponent(tokenToFetch)}`, token);
  },

  async pruneToken(token: string, tokenToPrune: string): Promise<void> {
    await request<void>('DELETE', `/prune/${encodeURIComponent(tokenToPrune)}`, token);
  },

  async createProduct(token: string, name: string): Promise<void> {
    await request<void>('PUT', '/product/create', token, { name });
  },

  async fetchProduct(token: string, productName: string): Promise<Product> {
    return request<Product>('GET', `/product/fetch/${encodeURIComponent(productName)}`, token);
  },

  async fetchAllProducts(token: string): Promise<Product[]> {
    return request<Product[]>('GET', '/product/fetch', token);
  },

  async deleteProduct(token: string, productName: string): Promise<void> {
    await request<void>('DELETE', `/product/delete/${encodeURIComponent(productName)}`, token);
  },

  async uploadVersion(token: string, product: string, version: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append('product', product);
    formData.append('version', version);
    formData.append('file', file);

    const response = await fetch(`${API_BASE}/product/upload`, {
      method: 'POST',
      headers: {
        'X-Api-Token': token,
      },
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({})) as ErrorResponse;
      throw new ApiError(response.status, errorData.error || 'Failed to upload version');
    }
  },

  async downloadVersion(token: string, product: string, version: string): Promise<Blob> {
    const response = await fetch(
      `${API_BASE}/product/download/${encodeURIComponent(product)}/${encodeURIComponent(version)}`,
      {
        method: 'GET',
        headers: {
          'X-Api-Token': token,
        },
      }
    );

    if (!response.ok) {
      throw new ApiError(response.status, 'Failed to download version');
    }

    return response.blob();
  },

  async deleteVersion(token: string, product: string, version: string): Promise<void> {
    await request<void>(
      'DELETE',
      `/product/delete/${encodeURIComponent(product)}/${encodeURIComponent(version)}`,
      token
    );
  },

  async addProductToken(
    token: string,
    product: string,
    tokenToAdd: string,
    permissions: TokenPermissions
  ): Promise<void> {
    await request<void>('PUT', '/product/modify/addToken', token, {
      product,
      token: tokenToAdd,
      permissions,
    });
  },

  async deleteProductToken(token: string, product: string, tokenToDelete: string): Promise<void> {
    await request<void>('DELETE', '/product/modify/deleteToken', token, {
      product,
      token: tokenToDelete,
    });
  },
};

export type { ValidateResponse, ApiToken, Product, TokenPermissions, Version, HealthResponse };
export { ApiError };
