package server

import (
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Web server for static pages and api",
	Long:  `This will run the web server`,
	Run:   serverRun,
}

func init() {
	ServerCmd.Flags().Int("http_port", 8080, "HTTP Port for cyboard used for redirect")
	ServerCmd.Flags().Int("https_port", 1443, "HTTPS Port for cyboard")
	viper.BindPFlag("server.http_port", ServerCmd.Flags().Lookup("http_port"))
	viper.BindPFlag("server.https_port", ServerCmd.Flags().Lookup("https_port"))
}

func serverRun(cmd *cobra.Command, args []string) {
	// MongoDB setup
	CreateUniqueIndexes()
	// Web Server Setup
	CreateStore()
	webRouter := CreateWebRouter()
	teamRouter := CreateTeamRouter()

	app := negroni.New()
	webRouter.PathPrefix("/").Handler(negroni.New(
		negroni.HandlerFunc(RequireLogin),
		negroni.Wrap(teamRouter),
	))
	app.Use(negroni.NewLogger())
	app.Use(negroni.NewStatic(http.Dir("static")))
	app.Use(negroni.HandlerFunc(CheckSessionID))

	app.UseHandler(webRouter)

	l := log.New(os.Stdout, "[negroni] ", 0)
	http_port := viper.GetString("server.http_port")
	https_port := viper.GetString("server.https_port")
	cert := viper.GetString("server.cert")
	key := viper.GetString("server.key")

	l.Printf("Server running at: https://%s:%s", viper.GetString("server.ip"), https_port)

	go http.ListenAndServe(":"+http_port, http.HandlerFunc(redir))

	l.Fatal(http.ListenAndServeTLS(":"+https_port, cert, key, app))
}

func redir(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+viper.GetString("server.ip")+":"+viper.GetString("https_port")+r.RequestURI, http.StatusMovedPermanently)
}
