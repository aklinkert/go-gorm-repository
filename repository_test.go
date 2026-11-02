package gormrepository

import (
	"errors"
	"testing"

	"github.com/aklinkert/go-logging"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type author struct {
	ID    string `gorm:"primaryKey"`
	Name  string
	Posts []post
}

type post struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	Category  string
	AuthorID  string
	Author    *author
	DeletedAt gorm.DeletedAt
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Every test gets its own in-memory database.
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("getting sql.DB: %v", err)
	}
	// A second connection would see a different, empty in-memory database.
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(&author{}, &post{}); err != nil {
		t.Fatalf("migrating: %v", err)
	}

	return db
}

// seed creates two authors with three posts, one of them soft deleted.
func seed(t *testing.T, repo Repository) {
	t.Helper()

	records := []interface{}{
		&author{ID: "a1", Name: "Alice"},
		&author{ID: "a2", Name: "Bob"},
		&post{ID: "p1", Title: "First", Category: "go", AuthorID: "a1"},
		&post{ID: "p2", Title: "Second", Category: "go", AuthorID: "a1"},
		&post{ID: "p3", Title: "Third", Category: "sql", AuthorID: "a2"},
	}
	for _, r := range records {
		if err := repo.Create(r); err != nil {
			t.Fatalf("seeding %+v: %v", r, err)
		}
	}

	if err := repo.Delete(&post{ID: "p3"}); err != nil {
		t.Fatalf("soft deleting: %v", err)
	}
}

func newSeededRepo(t *testing.T, defaultJoins ...string) TransactionRepository {
	t.Helper()

	repo := NewGormRepository(newTestDB(t), logging.NewTestLogger(t), defaultJoins...)
	seed(t, repo)

	return repo
}

func ids(posts []post) []string {
	out := make([]string, 0, len(posts))
	for _, p := range posts {
		out = append(out, p.ID)
	}

	return out
}

func TestGetMany(t *testing.T) {
	tests := []struct {
		name  string
		query func(Repository, *[]post) error
		want  []string
	}{
		{
			name:  "GetAll includes soft deleted rows",
			query: func(r Repository, p *[]post) error { return r.GetAll(p) },
			want:  []string{"p1", "p2", "p3"},
		},
		{
			name:  "GetBatch",
			query: func(r Repository, p *[]post) error { return r.GetBatch(p, 2, 1) },
			want:  []string{"p2", "p3"},
		},
		{
			name:  "GetWhere",
			query: func(r Repository, p *[]post) error { return r.GetWhere(p, "category = 'go'") },
			want:  []string{"p1", "p2"},
		},
		{
			name:  "GetWhere excludes soft deleted rows",
			query: func(r Repository, p *[]post) error { return r.GetWhere(p, "category = 'sql'") },
			want:  []string{},
		},
		{
			name:  "GetWhereBatch",
			query: func(r Repository, p *[]post) error { return r.GetWhereBatch(p, "category = 'go'", 1, 1) },
			want:  []string{"p2"},
		},
		{
			name:  "GetByField",
			query: func(r Repository, p *[]post) error { return r.GetByField(p, "author_id", "a1") },
			want:  []string{"p1", "p2"},
		},
		{
			name: "GetByFields",
			query: func(r Repository, p *[]post) error {
				return r.GetByFields(p, map[string]interface{}{"author_id": "a1", "title": "Second"})
			},
			want: []string{"p2"},
		},
		{
			name:  "GetByFieldBatch",
			query: func(r Repository, p *[]post) error { return r.GetByFieldBatch(p, "category", "go", 1, 0) },
			want:  []string{"p1"},
		},
		{
			name: "GetByFieldsBatch",
			query: func(r Repository, p *[]post) error {
				return r.GetByFieldsBatch(p, map[string]interface{}{"category": "go"}, 5, 1)
			},
			want: []string{"p2"},
		},
		{
			name:  "no matches is not an error",
			query: func(r Repository, p *[]post) error { return r.GetByField(p, "author_id", "unknown") },
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSeededRepo(t)

			var posts []post
			if err := tt.query(repo, &posts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := ids(posts)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetOne(t *testing.T) {
	tests := []struct {
		name    string
		query   func(Repository, *post) error
		want    string
		wantErr error
	}{
		{
			name:  "GetOneByID",
			query: func(r Repository, p *post) error { return r.GetOneByID(p, "p2") },
			want:  "p2",
		},
		{
			name:    "GetOneByID not found",
			query:   func(r Repository, p *post) error { return r.GetOneByID(p, "missing") },
			wantErr: ErrNotFound,
		},
		{
			name:    "GetOneByID soft deleted",
			query:   func(r Repository, p *post) error { return r.GetOneByID(p, "p3") },
			wantErr: ErrNotFound,
		},
		{
			name:  "GetOneByField",
			query: func(r Repository, p *post) error { return r.GetOneByField(p, "title", "First") },
			want:  "p1",
		},
		{
			name:    "GetOneByField not found",
			query:   func(r Repository, p *post) error { return r.GetOneByField(p, "title", "Nope") },
			wantErr: ErrNotFound,
		},
		{
			name: "GetOneByFields",
			query: func(r Repository, p *post) error {
				return r.GetOneByFields(p, map[string]interface{}{"category": "go", "title": "Second"})
			},
			want: "p2",
		},
		{
			name: "GetOneByFields not found",
			query: func(r Repository, p *post) error {
				return r.GetOneByFields(p, map[string]interface{}{"category": "sql", "title": "Second"})
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSeededRepo(t)

			var p post
			err := tt.query(repo, &p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && p.ID != tt.want {
				t.Errorf("ID = %q, want %q", p.ID, tt.want)
			}
		})
	}
}

func TestPreloads(t *testing.T) {
	repo := newSeededRepo(t)

	var a author
	if err := repo.GetOneByID(&a, "a1", "Posts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := ids(a.Posts); len(got) != 2 {
		t.Errorf("preloaded posts = %v, want p1 and p2", got)
	}
}

func TestDefaultJoins(t *testing.T) {
	repo := newSeededRepo(t, "Author")

	var p post
	if err := repo.GetOneByID(&p, "p1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Author == nil || p.Author.Name != "Alice" {
		t.Errorf("joined author = %+v, want Alice", p.Author)
	}
}

func TestSaveAndDelete(t *testing.T) {
	repo := newSeededRepo(t)

	p := post{ID: "p1", Title: "Renamed", Category: "go", AuthorID: "a1"}
	if err := repo.Save(&p); err != nil {
		t.Fatalf("saving: %v", err)
	}

	var got post
	if err := repo.GetOneByID(&got, "p1"); err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if got.Title != "Renamed" {
		t.Errorf("title = %q, want %q", got.Title, "Renamed")
	}

	if err := repo.Delete(&got); err != nil {
		t.Fatalf("deleting: %v", err)
	}
	if err := repo.GetOneByID(&got, "p1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("error after delete = %v, want %v", err, ErrNotFound)
	}
}

func TestTransactions(t *testing.T) {
	errRollback := errors.New("rollback")

	tests := []struct {
		name      string
		commit    bool
		wantP4    bool
		wantTitle string
	}{
		{name: "commit", commit: true, wantP4: true, wantTitle: "Saved in tx"},
		{name: "rollback", commit: false, wantP4: false, wantTitle: "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSeededRepo(t)

			err := repo.DB().Transaction(func(tx *gorm.DB) error {
				if err := repo.CreateTx(&post{ID: "p4", Title: "Created in tx", AuthorID: "a2"}, tx); err != nil {
					return err
				}
				if err := repo.SaveTx(&post{ID: "p1", Title: "Saved in tx", Category: "go", AuthorID: "a1"}, tx); err != nil {
					return err
				}
				if err := repo.DeleteTx(&post{ID: "p4"}, tx); err != nil {
					return err
				}
				if !tt.commit {
					return errRollback
				}
				return nil
			})
			if tt.commit && err != nil {
				t.Fatalf("transaction failed: %v", err)
			}
			if !tt.commit && !errors.Is(err, errRollback) {
				t.Fatalf("transaction error = %v, want %v", err, errRollback)
			}

			var p1 post
			if err := repo.GetOneByID(&p1, "p1"); err != nil {
				t.Fatalf("reloading p1: %v", err)
			}
			if p1.Title != tt.wantTitle {
				t.Errorf("p1 title = %q, want %q", p1.Title, tt.wantTitle)
			}

			// p4 is soft deleted when committed and absent when rolled back,
			// so only the unscoped GetAll sees it, and only after a commit.
			var all []post
			if err := repo.GetAll(&all); err != nil {
				t.Fatalf("listing: %v", err)
			}
			found := false
			for _, p := range all {
				found = found || p.ID == "p4"
			}
			if found != tt.wantP4 {
				t.Errorf("p4 present unscoped = %v, want %v", found, tt.wantP4)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	log := logging.NewNonFailingTestLogger(t)
	repo := NewGormRepository(newTestDB(t), log)

	var posts []post
	err := repo.GetWhere(&posts, "no_such_column = 1")
	if err == nil {
		t.Fatal("expected an error for an invalid query")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want a database error", err)
	}
	if log.Errors() != 1 {
		t.Errorf("logged errors = %d, want 1", log.Errors())
	}

	// A not found error from gorm is not an error for list queries.
	if err := repo.HandleError(&gorm.DB{Error: gorm.ErrRecordNotFound}); err != nil {
		t.Errorf("HandleError(ErrRecordNotFound) = %v, want nil", err)
	}
	if err := repo.HandleOneError(&gorm.DB{Error: gorm.ErrRecordNotFound}); !errors.Is(err, ErrNotFound) {
		t.Errorf("HandleOneError(ErrRecordNotFound) = %v, want %v", err, ErrNotFound)
	}

	errBoom := errors.New("boom")
	if err := repo.HandleOneError(&gorm.DB{Error: errBoom}); !errors.Is(err, errBoom) {
		t.Errorf("HandleOneError(boom) = %v, want it to wrap %v", err, errBoom)
	}
}
