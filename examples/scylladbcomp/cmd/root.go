package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"github.com/taimaifika/service-context/component/scylladbc"
	"github.com/taimaifika/service-context/examples/scylladbcomp/repo"

	sctx "github.com/taimaifika/service-context"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithComponent(scylladbc.NewScyllaDbComponent("scylladb")),
	)
}

type ScyllaDbComponent interface {
	GetSession() *gocql.Session
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start mongodb service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("load service context error", "error", err)
			panic(err)
		}

		// Get ScyllaDB component
		scyllaDbComponent := serviceCtx.MustGet("scylladb").(ScyllaDbComponent)
		session := scyllaDbComponent.GetSession()

		if session == nil {
			slog.Error("failed to get ScyllaDB session")
			return
		}

		// New ScyllaDbRepo instance
		scyllaDbRepo := repo.NewScyllaDbRepo(session)
		// Example: Create a table
		if err := scyllaDbRepo.CreateExampleTable(); err != nil {
			slog.Error("failed to create example table", "error", err)
			return
		}
		slog.Info("Example table created successfully")

		// Example: Insert a user
		id := gocql.TimeUUID()
		if err := scyllaDbRepo.InsertUser(id, "John Doe", "john@example.com"); err != nil {
			slog.Error("failed to insert user", "error", err)
			return
		}
		slog.Info("User inserted successfully", "id", id)

		// Example: Get a user
		name, email, createdAt, err := scyllaDbRepo.GetUser(id)
		if err != nil {
			slog.Error("failed to get user", "error", err)
			return
		}
		slog.Info("User retrieved successfully", "id", id, "name", name, "email", email, "created_at", createdAt)

		// Example: Get all users
		users, err := scyllaDbRepo.GetAllUsers()
		if err != nil {
			slog.Error("failed to get all users", "error", err)
			return
		}
		slog.Info("All users retrieved successfully", "users", users)

		// Example: Delete a user
		if err := scyllaDbRepo.DeleteUser(id); err != nil {
			slog.Error("failed to delete user", "error", err)
			return
		}
		slog.Info("User deleted successfully", "id", id)

		// Example: Update a user
		if err := scyllaDbRepo.UpdateUser(id, "Jane Doe", "john@example.com"); err != nil {
			slog.Error("failed to update user", "error", err)
			return
		}
		slog.Info("User updated successfully", "id", id)

		// Example: Count users
		count, err := scyllaDbRepo.CountUsers()
		if err != nil {
			slog.Error("failed to count users", "error", err)
			return
		}
		slog.Info("Total users count", "count", count)

		// Example: Get users with pagination
		paginatedUsers, err := scyllaDbRepo.GetUsersWithPagination(0, 10)
		if err != nil {
			slog.Error("failed to get users with pagination", "error", err)
			return
		}
		slog.Info("Paginated users retrieved successfully", "users", paginatedUsers)

		// Example: Use prepared statements
		if err := scyllaDbRepo.PreparedInsertUser(); err != nil {
			slog.Error("failed to use prepared statement for user insertion", "error", err)
			return
		}
		slog.Info("Prepared statement for user insertion executed successfully")

		// Example: Batch insert users
		usersToInsert := []repo.User{
			{ID: gocql.TimeUUID(), Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()},
			{ID: gocql.TimeUUID(), Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()},
		}
		if err := scyllaDbRepo.BatchInsertUsers(usersToInsert); err != nil {
			slog.Error("failed to batch insert users", "error", err)
			return
		}
		slog.Info("Batch insert of users executed successfully", "users", usersToInsert)

		// Done with all operations
		slog.Info("All ScyllaDB operations completed successfully")
	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
