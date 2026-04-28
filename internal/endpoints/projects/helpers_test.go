package projects

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"packster/internal"
	"packster/pkg/config"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func newCtx(t *testing.T, method, target string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var b *bytes.Buffer
	if body != nil {
		b = bytes.NewBuffer(body)
	} else {
		b = &bytes.Buffer{}
	}
	c.Request = httptest.NewRequest(method, target, b)
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

func signSession(t *testing.T, userID int, hostURL string, orgs []int) string {
	t.Helper()
	if orgs == nil {
		orgs = []int{}
	}
	claims := jwt.MapClaims{
		"sub":   strconv.Itoa(userID),
		"token": "provider-token",
		"host":  map[string]string{"type": "gitlab", "url": hostURL},
		"orgs":  orgs,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

func setAuthHeader(c *gin.Context, jwt string) {
	c.Request.Header.Set("Authorization", "Bearer "+jwt)
}

func withHosts(hosts map[string]types.Host) func() {
	prev := internal.Hosts
	internal.Hosts = hosts
	return func() { internal.Hosts = prev }
}

// ---- fake repos ----

type fakeUserRepo struct {
	existsFn   func(id int) (bool, error)
	searchFn   func(hostID int, q string, exclude int) ([]types.User, error)
}

func (f *fakeUserRepo) CreateUser(string, string, int, []int) (*types.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) UserExists(string, int, int) (*types.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) UserExistsByID(id int) (bool, error) {
	if f.existsFn != nil {
		return f.existsFn(id)
	}
	return true, nil
}
func (f *fakeUserRepo) PurgeUserData(int) ([]string, error) {
	return nil, nil
}
func (f *fakeUserRepo) SearchByName(hostID int, q string, exclude int) ([]types.User, error) {
	if f.searchFn != nil {
		return f.searchFn(hostID, q, exclude)
	}
	return nil, nil
}

type fakeProjectRepo struct {
	getByIDFn        func(id int) (*types.Project, error)
	listAccessibleFn func(userID int) ([]types.Project, error)
	getByHRFn        func(host, repo int) (*types.Project, error)
	importFn         func(ownerID, hostID, repo int) (*types.Project, error)
	deleteFn         func(id int) ([]string, error)
}

func (f *fakeProjectRepo) Import(o, h, r int) (*types.Project, error) {
	if f.importFn != nil {
		return f.importFn(o, h, r)
	}
	return nil, nil
}
func (f *fakeProjectRepo) GetByID(id int) (*types.Project, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(id)
	}
	return nil, nil
}
func (f *fakeProjectRepo) ListAccessible(userID int) ([]types.Project, error) {
	if f.listAccessibleFn != nil {
		return f.listAccessibleFn(userID)
	}
	return nil, nil
}
func (f *fakeProjectRepo) GetByHostRepository(h, r int) (*types.Project, error) {
	if f.getByHRFn != nil {
		return f.getByHRFn(h, r)
	}
	return nil, nil
}
func (f *fakeProjectRepo) Delete(id int) ([]string, error) {
	if f.deleteFn != nil {
		return f.deleteFn(id)
	}
	return nil, nil
}

type fakePermissionRepo struct {
	getFn    func(userID, projectID int) (*types.Permission, error)
	setFn    func(p types.Permission) error
	deleteFn func(userID, projectID int) error
	listFn   func(projectID int) ([]types.PermissionEntry, error)

	setCalls    []types.Permission
	deleteCalls [][2]int
}

func (f *fakePermissionRepo) Get(u, p int) (*types.Permission, error) {
	if f.getFn != nil {
		return f.getFn(u, p)
	}
	return nil, nil
}
func (f *fakePermissionRepo) Set(p types.Permission) error {
	f.setCalls = append(f.setCalls, p)
	if f.setFn != nil {
		return f.setFn(p)
	}
	return nil
}
func (f *fakePermissionRepo) Delete(u, p int) error {
	f.deleteCalls = append(f.deleteCalls, [2]int{u, p})
	if f.deleteFn != nil {
		return f.deleteFn(u, p)
	}
	return nil
}
func (f *fakePermissionRepo) ListByProject(projectID int) ([]types.PermissionEntry, error) {
	if f.listFn != nil {
		return f.listFn(projectID)
	}
	return nil, nil
}

type fakeProductRepo struct {
	createFn func(projectID int, name string) (*types.Product, error)
	getByIDFn func(id int) (*types.Product, error)
	getByNameFn func(projectID int, name string) (*types.Product, error)
	listFn   func(projectID int) ([]types.Product, error)
	deleteFn func(id int) error
}

func (f *fakeProductRepo) Create(projectID int, name string) (*types.Product, error) {
	if f.createFn != nil {
		return f.createFn(projectID, name)
	}
	return nil, nil
}
func (f *fakeProductRepo) GetByID(id int) (*types.Product, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(id)
	}
	return nil, nil
}
func (f *fakeProductRepo) GetByName(projectID int, name string) (*types.Product, error) {
	if f.getByNameFn != nil {
		return f.getByNameFn(projectID, name)
	}
	return nil, nil
}
func (f *fakeProductRepo) ListByProject(projectID int) ([]types.Product, error) {
	if f.listFn != nil {
		return f.listFn(projectID)
	}
	return nil, nil
}
func (f *fakeProductRepo) Delete(id int) error {
	if f.deleteFn != nil {
		return f.deleteFn(id)
	}
	return nil
}

type fakeVersionRepo struct {
	createFn  func(productID int, name, path, checksum string) (*types.Version, error)
	getByIDFn func(id int) (*types.Version, error)
	getByNameFn func(productID int, name string) (*types.Version, error)
	listFn    func(productID int) ([]types.Version, error)
	deleteFn  func(id int) error
}

func (f *fakeVersionRepo) Create(productID int, name, path, checksum string) (*types.Version, error) {
	if f.createFn != nil {
		return f.createFn(productID, name, path, checksum)
	}
	return nil, nil
}
func (f *fakeVersionRepo) GetByID(id int) (*types.Version, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(id)
	}
	return nil, nil
}
func (f *fakeVersionRepo) GetByName(productID int, name string) (*types.Version, error) {
	if f.getByNameFn != nil {
		return f.getByNameFn(productID, name)
	}
	return nil, nil
}
func (f *fakeVersionRepo) ListByProduct(productID int) ([]types.Version, error) {
	if f.listFn != nil {
		return f.listFn(productID)
	}
	return nil, nil
}
func (f *fakeVersionRepo) Delete(id int) error {
	if f.deleteFn != nil {
		return f.deleteFn(id)
	}
	return nil
}

func newHandler(
	user *fakeUserRepo,
	project *fakeProjectRepo,
	perm *fakePermissionRepo,
	product *fakeProductRepo,
	version *fakeVersionRepo,
) *ProjectsHandler {
	return &ProjectsHandler{
		Cfg:            testCfg(),
		UserRepo:       user,
		ProjectRepo:    project,
		PermissionRepo: perm,
		ProductRepo:    product,
		VersionRepo:    version,
		HTTP:           &http.Client{},
	}
}

func testCfg() config.Config {
	return config.Config{Secret: testSecret}
}
