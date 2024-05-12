package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

type Article struct {
	Id                   uint16
	Title, Idea, Article string
}

var posts = []Article{}
var watchPost = Article{}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("C:/www/templates/index.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM articles")

	if err != nil {
		panic(err)
	}

	posts = []Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Idea, &post.Article)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	t.ExecuteTemplate(w, "index", posts)
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("C:/www/templates/create.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "create", nil)
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	idea := r.FormValue("idea")
	article := r.FormValue("article")

	// if title == "" || idea == "" || article == "" {
	// 	fmt.Fprintf(w, "Пожалуйста, заполните все поля")
	// } else {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `idea`, `aricle`) VALUES ('%s', '%s', '%s')", title, idea, article))

	if err != nil {
		panic(err)
	}

	defer insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func watch_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("C:/www/templates/watch.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query(fmt.Sprintf("SELECT * FROM articles WHERE id = '%s'", vars["id"]))

	if err != nil {
		panic(err)
	}

	watchPost = Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Idea, &post.Article)
		if err != nil {
			panic(err)
		}

		watchPost = post
	}

	t.ExecuteTemplate(w, "watch", watchPost)
}

func search(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM articles WHERE title LIKE ? OR idea LIKE ? OR aricle LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		panic(err)
	}

	searchResults := []Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Idea, &post.Article)
		if err != nil {
			panic(err)
		}

		searchResults = append(searchResults, post)
	}

	t, err := template.ParseFiles("C:/www/templates/search.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")

	t.ExecuteTemplate(w, "search", searchResults)
}

func allArticles(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("C:/www/templates/allArticles.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM articles")

	if err != nil {
		panic(err)
	}

	allPosts := []Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Idea, &post.Article)
		if err != nil {
			panic(err)
		}

		allPosts = append(allPosts, post)
	}

	t.ExecuteTemplate(w, "allArticles", allPosts)
}

type User struct {
	ID       int
	Username string
	Password string
}

func registerUser(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username != "" && password != "" {

			hashedPassword, err := hashPassword(password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer db.Close()

			// _, err = db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", username, hashedPassword)
			_, err = db.Query(fmt.Sprintf("INSERT INTO `user` (`username`, `password`) VALUES ('%s', '%s')", username, hashedPassword))

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			http.Error(w, "Заполните все поля", http.StatusBadRequest)
		}
	} else {
		t, err := template.ParseFiles("C:/www/templates/register.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "register", nil)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

var store = sessions.NewCookieStore([]byte("your-secret-key"))

func loginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/golang")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var user User
		err = db.QueryRow("SELECT id, username, password FROM user WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
			return
		}

		err = checkPasswordHash(password, user.Password)
		if err != nil {
			http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
			return
		}

		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["user_id"] = user.ID
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {

		t, err := template.ParseFiles("C:/www/templates/login.html", "C:/www/templates/header.html", "C:/www/templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "login", nil)

	}
}

func handleFunc() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", watch_post).Methods("GET")
	rtr.HandleFunc("/search", search).Methods("GET")
	rtr.HandleFunc("/all-articles", allArticles).Methods("GET")
	http.Handle("/", rtr)
	rtr.HandleFunc("/register", registerUser).Methods("POST", "GET")
	rtr.HandleFunc("/login", loginUser).Methods("GET", "POST")

	http.ListenAndServe(":8080", nil)
}

func main() {
	handleFunc()
}
