package internal

import (
	"fmt"
	"testing"

	"packster/pkg/config"
	"packster/pkg/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetHosts() {
	Hosts = nil
}

func TestFetchOrgsByHostUrl_HostNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("https://missing").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	orgs, id, err := FetchOrgsByHostUrl(db, "https://missing")
	assert.Error(t, err)
	assert.Nil(t, orgs)
	assert.Equal(t, -1, id)
	assert.Contains(t, err.Error(), "host not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchOrgsByHostUrl_ExistsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("https://gitlab.example.com").
		WillReturnError(fmt.Errorf("db down"))

	orgs, id, err := FetchOrgsByHostUrl(db, "https://gitlab.example.com")
	assert.Error(t, err)
	assert.Nil(t, orgs)
	assert.Equal(t, -1, id)
}

func TestFetchOrgsByHostUrl_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	url := "https://gitlab.example.com"

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT o.id, h.id`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"o.id", "h.id"}).
			AddRow(10, 1).
			AddRow(20, 1).
			AddRow(30, 1))

	orgs, id, err := FetchOrgsByHostUrl(db, url)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, []int{10, 20, 30}, orgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchOrgsByHostUrl_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	url := "https://gitlab.example.com"

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT o.id, h.id`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"o.id", "h.id"}))

	orgs, id, err := FetchOrgsByHostUrl(db, url)
	assert.NoError(t, err)
	assert.Equal(t, 0, id)
	assert.Nil(t, orgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchOrgsByHostUrl_SelectError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	url := "https://gitlab.example.com"

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT o.id, h.id`).
		WithArgs(url).
		WillReturnError(fmt.Errorf("query failed"))

	orgs, id, err := FetchOrgsByHostUrl(db, url)
	assert.Error(t, err)
	assert.Nil(t, orgs)
	assert.Equal(t, -1, id)
}

func TestFetchHosts_NilConn(t *testing.T) {
	defer resetHosts()
	cfg := config.Config{}
	err := FetchHosts(cfg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Sql conn is nil")
}

func TestFetchHosts_InitializesMap(t *testing.T) {
	defer resetHosts()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	Hosts = nil
	cfg := config.Config{}
	err = FetchHosts(cfg, db)
	assert.NoError(t, err)
	assert.NotNil(t, Hosts)
	assert.Empty(t, Hosts)
}

func TestFetchHosts_NilGitlabConfig(t *testing.T) {
	defer resetHosts()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	Hosts = make(map[string]types.Host)
	cfg := config.Config{Gitlab: nil}
	err = FetchHosts(cfg, db)
	assert.NoError(t, err)
	assert.Empty(t, Hosts)
}

func TestFetchHosts_LoadsGitlabHosts(t *testing.T) {
	defer resetHosts()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	Hosts = make(map[string]types.Host)

	url := "https://gitlab.example.com"
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT o.id, h.id`).
		WithArgs(url).
		WillReturnRows(sqlmock.NewRows([]string{"o.id", "h.id"}).
			AddRow(10, 2).
			AddRow(11, 2))

	cfg := config.Config{
		Gitlab: &config.GitlabConfig{
			Hosts: map[string]config.GitlabHost{
				url: {ApplicationId: "app", Secret: "sec"},
			},
		},
	}

	err = FetchHosts(cfg, db)
	assert.NoError(t, err)
	require.Contains(t, Hosts, url)
	h := Hosts[url]
	assert.Equal(t, 2, h.Id)
	assert.Equal(t, url, h.Url)
	assert.Equal(t, types.Gitlab, h.Type)
	assert.Equal(t, "app", h.ApplicationId)
	assert.Equal(t, "sec", h.Secret)
	assert.Equal(t, []int{10, 11}, h.Orgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchHosts_SkipsHostOnError(t *testing.T) {
	defer resetHosts()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	Hosts = make(map[string]types.Host)

	goodURL := "https://good.example.com"
	badURL := "https://bad.example.com"

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(badURL).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(goodURL).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT o.id, h.id`).
		WithArgs(goodURL).
		WillReturnRows(sqlmock.NewRows([]string{"o.id", "h.id"}).AddRow(5, 1))

	cfg := config.Config{
		Gitlab: &config.GitlabConfig{
			Hosts: map[string]config.GitlabHost{
				goodURL: {ApplicationId: "a", Secret: "s"},
				badURL:  {ApplicationId: "b", Secret: "t"},
			},
		},
	}

	err = FetchHosts(cfg, db)
	assert.NoError(t, err)
	assert.Contains(t, Hosts, goodURL)
	assert.NotContains(t, Hosts, badURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}
