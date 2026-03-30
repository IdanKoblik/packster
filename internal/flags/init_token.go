package flags

import (
	"context"
	"packster/internal/logging"
	"packster/internal/repository"
	"packster/pkg/flags"
	"packster/pkg/types"
)

func InitToken(repo *repository.AuthRepository) flags.Flag {
	return flags.Flag{
		Cmd:  "--init-admin-token",
		Name: "init-admin-token",
		Args: []string{},
		Description: []string{
			"Creates initial token that is an admin token.",
			"Please remove this flag after initial use",
		},
		Handle: func(args []string) error {
			exists, err := adminTokenExists(repo)
			if err != nil {
				return err
			}

			if exists {
				logging.Log.Warn("Admin token already exists, please remove this flag")
				return nil
			}

			token, err := repo.CreateToken(&types.RegisterRequest{
				Admin: true,
			})

			if err != nil {
				return err
			}

			logging.Log.Infof("Initial token %s", token)
			return nil
		},
	}
}

func adminTokenExists(r *repository.AuthRepository) (bool, error) {
	var count int
	err := r.DB.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM principals WHERE admin = TRUE AND type = 'token'").Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
