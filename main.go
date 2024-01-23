package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath = "ma_base_de_donnees.db"
)
const (
	smtpServer   = "smtp.gmail.com"
	smtpPort     = 587
	smtpUsername = "camarachristine627@gmail.com"
	smtpPassword = "pxdw arja nolq lmig"
)
var db *sql.DB

// User structure to represent a user
type User struct {
	ID         int
	Nom        string
	Prenom     string
	Email      string
	Number     string
	Ville      string
	Motivation string
	Message string
}


func init() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	// Create users table if not exists
	/*_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS utilisateurs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            nom TEXT NOT NULL,
            prenom TEXT NOT NULL,
            email TEXT NOT NULL,
            number TEXT NOT NULL,
            ville TEXT NOT NULL,
            motivation TEXT NOT NULL
        )
    `)
	if err != nil {
		panic(err)
	}
	*/

	fmt.Println("Base de données initialisée avec succès")
}

func connexionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	nom := r.FormValue("nom")
	prenom := r.FormValue("prenom")
	email := r.FormValue("email")
	number := r.FormValue("number")
	ville := r.FormValue("ville")
	motivation := r.FormValue("motivation")
	message:=r.FormValue("message")

	// Insert user into the database
	_, err := db.Exec("INSERT INTO utilisateurs(nom, prenom, email, number, ville, motivation,message) VALUES (?, ?, ?, ?, ?, ?, ?)",
		nom, prenom, email, number, ville, motivation, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Utilisateur inséré avec succès :", nom)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	 // Envoi de l'e-mail de confirmation
	 subject := "CONFIRMATION D'ENREGISTREMENT"
	 body := "Votre enregistrement a été effectué avec succès. Merci de faire partie de notre communauté."
 
	 err = sendEmail(email, subject, body)
	 if err != nil {
		 // Gérer l'erreur de l'envoi de l'e-mail, par exemple, le journaliser
		 fmt.Println("Erreur lors de l'envoi de l'e-mail:", err)
	 }
}

func main() {
	r := mux.NewRouter()

	// Chargement des fichiers statiques depuis le dossier "static"
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/connexion", connexionHandler)

	// Serve les fichiers statiques du dossier "static"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Gère les requêtes avec la fonction `handler`
	http.HandleFunc("/", handler)
	http.HandleFunc("/connexion", connexionHandler)

	// Démarre le serveur sur le port 8000
	port := 8082
	fmt.Printf("Serveur en cours d'exécution sur le port %d...\n", port)
	fmt.Println("http://localhost:8082")
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Erreur :", err)
	}
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Fonction handler qui gère toutes les requêtes HTTP
func handler(w http.ResponseWriter, r *http.Request) {
	// Récupère le chemin demandé dans la requête
	path := r.URL.Path[1:]

	// Si le chemin est vide, renvoie index.html par défaut
	if path == "" {
		path = "index.html"
	}

	// Lit le fichier correspondant au chemin
	content, err := readFile(path)
	if err != nil {
		// Si le fichier n'est pas trouvé, renvoie une erreur 404
		http.NotFound(w, r)
		return
	}

	// Détermine le type de contenu en fonction de l'extension du fichier
	contentType := getContentType(path)

	// Définit le type de contenu dans l'en-tête de la réponse
	w.Header().Set("Content-Type", contentType)

	// Écrit le contenu du fichier dans la réponse
	w.Write(content)
}

// Fonction pour lire le contenu d'un fichier
func readFile(filename string) ([]byte, error) {
	path := fmt.Sprintf("./templates/%s", filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("le fichier %s n'existe pas", filename)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// Fonction pour obtenir le type de contenu en fonction de l'extension du fichier
func getContentType(filename string) string {
	// Vous pouvez étendre cette fonction pour gérer d'autres types de fichiers
	if filename[len(filename)-5:] == ".html" {
		return "text/html"
	} else if filename[len(filename)-4:] == ".css" {
		return "text/css"
	} else {
		return "text/plain"
	}
}
func sendEmail(to, subject, body string) error {
	// Crée le message au format MIME
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	// Établit une connexion avec le serveur SMTP
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpServer, smtpPort), auth, smtpUsername, []string{to}, []byte(message))
	if err != nil {
		return err
	}

	fmt.Println("E-mail envoyé avec succès à", to)
	return nil
}