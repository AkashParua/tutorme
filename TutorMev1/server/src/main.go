package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

     _ "github.com/mattn/go-sqlite3"

)

type User struct {
    ID       int    	`json:"id"`
    Username string 	`json:"username"`
    Password string 	`json:"password"`
    Email    string  	`json:"email"`
}

type Question struct {
    ID        int   		`json:"id"`
    Content   string 		`json:"content"`
    CreatedAt time.Time		`json:"created_at"`
    UserID    int			`json:"user_id"`
}

type Answer struct {
    ID         int			`json:"id"`
    Content    string 		`json:"content"`
    CreatedAt  time.Time    `json:"created_at"`
    QuestionID int  		`json:"question_id"`
    UserID     int 			`json:"user_id"`
    BookName   string 		`json:"book_name"`
    FileLink   string 		`json:"file_link"`
}

type Book struct {
    ID       int			`json:"id"`
    BookName string			`json:"book_name"`
    AddedOn  time.Time		`json:"added_on"`
    FileLink string 		`json:"file_link"`
    UserID   int			`json:"user_id"`
}

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

var jwtKey = []byte("")

func main() {
	// Loading the .env file
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	port := os.Getenv("PORT")
	jwtKey = []byte(os.Getenv("JWTKEY"))

    db, err := sql.Open("sqlite3" ,"database/my_db.sqlite")
    if err != nil {
		fmt.Println("Failed to connect to the database!")
        log.Fatal(err)
    }
    defer db.Close()
	//add command line message if the connection is successful
	fmt.Println("Successfully connected to the database!")

    router := mux.NewRouter()

    router.HandleFunc("/login", loginHandler).Methods("POST")
    router.HandleFunc("/questions", authMiddleware(getQuestionsHandler)).Methods("GET").Queries("userid" , "{userid}")
    router.HandleFunc("/answers", authMiddleware(getAnswersHandler)).Methods("GET").Queries("userid" , "{userid}")
    router.HandleFunc("/books", authMiddleware(getBooksHandler)).Methods("GET").Queries("userid" , "{userid}")
	router.HandleFunc("/questions", authMiddleware(addQuestionHandler)).Methods("POST")
	router.HandleFunc("/answers", authMiddleware(addAnswerHandler)).Methods("POST")
	router.HandleFunc("/books", authMiddleware(addBookHandler)).Methods("POST")
	router.HandleFunc("/signup", signupHandler).Methods("POST")
	router.HandleFunc("/bookbyname" ,getBookbyNameHandler).Methods("GET").Queries("bookname" , "{bookname}")
	log.Fatal(http.ListenAndServe(port, router))
}

//Signing in the handler function
func signupHandler(w http.ResponseWriter, r *http.Request) {
	
	// Retrieving the user information from the request body into user struct
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3" ,"database/my_db.sqlite")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//----------------------------------//////////////////SQL query to insert the user into the database/////////////////-----------------------
	_, err = db.Exec("INSERT INTO users (username, password, email) VALUES ($1, $2, $3)", user.Username, string(hashedPassword), user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
}

//login handler function
func loginHandler(w http.ResponseWriter, r *http.Request) {

	// Retrieve the user information from the request body into user struct
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Checking if the user exists in the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()
	// Checking the infomration of the user from the database
	var user_info User 
    row := db.QueryRow("SELECT id, username, password FROM users WHERE email = $1", user.Email)
    err = row.Scan(&user_info.ID, &user_info.Username, &user_info.Password, &user.Email)
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Comparing the hashed password with the password from the database
    err = bcrypt.CompareHashAndPassword([]byte(user_info.Password), []byte(user.Password))
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Generate a JWT token with the user's information
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": user.Username,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iss": "my_issuer",
    })

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return the JWT token to the client
    response := map[string]string{"token": tokenString}
    json.NewEncoder(w).Encode(response)
}

//====================================================================================================//
// Get all the questions from the database for a specific user
func getBookbyNameHandler(w http.ResponseWriter, r *http.Request) {
	
	bookname := mux.Vars(r)["bookname"]

	// Retrieve the questions from the database
	db, err := sql.Open("sqlite3","database/my_db.sqlite")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	rows, err := db.Query("SELECT id, book_name, added_on, file_link, user_id FROM books WHERE book_name = $1 ORDER BY added_on",bookname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.BookName, &book.AddedOn, &book.FileLink, &book.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	// Return the questions to the client
	json.NewEncoder(w).Encode(books)
}
func getQuestionsHandler(w http.ResponseWriter, r *http.Request) {

	userid := mux.Vars(r)["userid"]

    // Retrieve the questions from the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()
	
    rows, err := db.Query("SELECT id, content, created_at, user_id FROM questions WHERE user_id = $1 ORDER BY created_at",userid)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var questions []Question
    for rows.Next() {
        var question Question
        err := rows.Scan(&question.ID, &question.Content, &question.CreatedAt, &question.UserID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        questions = append(questions, question)
    }

    // Return the questions to the client
    json.NewEncoder(w).Encode(questions)
}
//--------------------------------------------------------------------------------------------
func getAnswersHandler(w http.ResponseWriter, r *http.Request) {
	userid := mux.Vars(r)["userid"]
    // Retrieve the answers from the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    rows, err := db.Query("SELECT id, content, created_at, question_id, user_id, book_name, file_link FROM answers WHERE user_id = $1 ORDER BY created_at", userid)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var answers []Answer
    for rows.Next() {
        var answer Answer
        err := rows.Scan(&answer.ID, &answer.Content, &answer.CreatedAt, &answer.QuestionID, &answer.UserID, &answer.BookName, &answer.FileLink)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        answers = append(answers, answer)
    }

    // Return the answers to the client
    json.NewEncoder(w).Encode(answers)
}
//-------------------------------------------------------------------------------------------
func getBooksHandler(w http.ResponseWriter, r *http.Request) {
	userid := mux.Vars(r)["userid"]

    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    rows , err := db.Query("SELECT * FROM books WHERE user_id = $1", userid)
    var books []Book
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.BookName, &book.AddedOn, &book.FileLink, &book.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}
	// Returning the books to the client
	json.NewEncoder(w).Encode(books)

}//=====================================================================================================================

// Middleware function to check if the JWT token is valid
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {

    return func(w http.ResponseWriter, r *http.Request) {
        // Get the JWT token from the Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        tokenString := authHeader[len("Bearer "):]
        // Parsing the JWT token
        token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return jwtKey, nil
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }
        // Verify the JWT token
        if !token.Valid {
            http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
            return
        }
        // Call the next handler function
        next(w, r)
    }
}

//=========================================================================================================================
// Add a new question to the database
func addQuestionHandler(w http.ResponseWriter, r *http.Request) {

    // Parse the question data from the request body
    var question Question
    err := json.NewDecoder(r.Body).Decode(&question)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Insert the question into the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO questions (id, content, user_id) VALUES ($1, $2, $3)", question.ID, question.Content, question.UserID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return a success response
    w.WriteHeader(http.StatusCreated)
}

// Add a new answer to the database
func addAnswerHandler(w http.ResponseWriter, r *http.Request) {

    // Parse the answer data from the request body
    var answer Answer
    err := json.NewDecoder(r.Body).Decode(&answer)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Insert the answer into the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO answers (content, question_id, user_id ,book_name , file_link) VALUES ($1, $2, $3, $4, $5)", answer.Content, answer.QuestionID, answer.UserID , answer.BookName , answer.FileLink)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return a success response
    w.WriteHeader(http.StatusCreated)
}

// Add a new book to the database for a specific user
func addBookHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the book data from the request body
    var book Book
    err := json.NewDecoder(r.Body).Decode(&book)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Handle the upload of the book file
    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Save the file to disk
    f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer f.Close()

    _, err = io.Copy(f, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Insert the book into the database
    db, err := sql.Open("sqlite3","database/my_db.sqlite")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO books (id , book_name, user_id, file_link) VALUES ($1, $2, $3)", book.ID ,book.BookName , book.UserID, handler.Filename)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return a success response
    w.WriteHeader(http.StatusCreated)
}