package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"mdu/mangasrc"
)

func NewMangadexCmd() *cobra.Command {
	mangadexCmd := &cobra.Command{
		Use:   "mangadex",
		Short: "MangaDex authentication",
		Long:  `Login/logout to MangaDex to access adult/Hentai content.`,
	}

	mangadexCmd.AddCommand(
		newMangadexLoginCmd(),
		newMangadexLogoutCmd(),
		newMangadexStatusCmd(),
	)

	return mangadexCmd
}

func newMangadexLoginCmd() *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to MangaDex",
		RunE: func(cmd *cobra.Command, args []string) error {
			if username == "" {
				fmt.Print("Username: ")
				fmt.Scanln(&username)
			}
			if password == "" {
				fmt.Print("Password: ")
				// simple read
				var p string
				fmt.Scanln(&p)
				password = p
			}
			cfg, err := mangasrc.Login(username, password)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Logged in as %s (session saved to ~/.config/mdu/mangadex_auth.json)\n", cfg.Username)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "MangaDex username")
	cmd.Flags().StringVar(&password, "password", "", "MangaDex password")

	return cmd
}

func newMangadexLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from MangaDex",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := mangasrc.Logout(); err != nil {
				return err
			}
			fmt.Println("✓ Logged out (credentials removed)")
			return nil
		},
	}
	return cmd
}

func newMangadexStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show MangaDex auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := mangasrc.LoadAuthConfig()
			if err != nil {
				return err
			}
			if cfg == nil || cfg.SessionToken == "" {
				fmt.Println("Not logged in")
				return nil
			}
			if cfg.Username != "" {
				fmt.Printf("Logged in as %s\n", cfg.Username)
			} else {
				fmt.Println("Logged in (session token present)")
			}
			if !cfg.SavedAt.IsZero() {
				fmt.Printf("Saved: %s\n", cfg.SavedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	return cmd
}
