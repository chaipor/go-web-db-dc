package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

const (
	host     = "pg-db"
	port     = 5432
	user     = "myuser"
	password = "mypassword"
	dbname   = "mytestdb"
)

type AllData struct {
	Item     int
	Name     string
	Nickname string
	Research string
	EditLink string
}

type addeditform_data struct {
	Title              string
	FormAction         string
	InputNameValue     string
	InputNameAttr      string
	InputNicknameValue string
	InputResearchValue string
	DisplayDelButton   string
}

var db *sql.DB

func ShowPageNotFound(w http.ResponseWriter, r *http.Request) {
	fmt.Println("404:" + r.URL.Path)
	tmpl := template.Must(template.ParseFiles("404.html"))
	tmpl.Execute(w, nil)
}

func ShowHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ShowPageNotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func ShowAboutPage(w http.ResponseWriter, r *http.Request) {
	if !(r.URL.Path == "/about" || r.URL.Path == "/about/") {
		ShowPageNotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles("about.html"))
	tmpl.Execute(w, nil)
}

func ShowUpdate(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	para1 := parts[2]
	var name, nickname, research string
	err := db.QueryRow("SELECT tb_Name.cName, tb_Name.cNickname, tb_Research.cResearch FROM tb_Name Left Join tb_Research on tb_Research.cName=tb_Name.cName WHERE tb_Name.cName=$1", para1).Scan(&name, &nickname, &research)

	if err != nil {
		http.Error(w, "Database query error", http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}

	data := addeditform_data{
		Title:              "Edit student research information",
		FormAction:         "/UpdateUserData",
		InputNameValue:     name,
		InputNameAttr:      "readonly",
		InputNicknameValue: nickname,
		InputResearchValue: research,
		DisplayDelButton:   "inline",
	}
	tmpl := template.Must(template.ParseFiles("addeditdata.html"))
	tmpl.Execute(w, data)
}

func WriteHttpHeaderResponse(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	fmt.Fprint(w, m)
}

func InsertUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Received POST request for insert data")
		var BadRequest bool = false

		name := r.FormValue("name")
		nickname := r.FormValue("nickname")
		research := r.FormValue("research")

		// Insert data into tbName table
		_, err := db.Exec("INSERT INTO tb_Name (cName, cNickname) VALUES ($1, $2)", name, nickname)
		if err != nil {
			fmt.Println(err)
			BadRequest = true
		} else {
			// Insert data into tbResearch table
			_, err = db.Exec("INSERT INTO tb_Research (cName, cResearch) VALUES ($1, $2)", name, research)
			if err != nil {
				fmt.Println(err)
				BadRequest = true
			}
		}

		// Error from inserting to database
		if BadRequest {
			WriteHttpHeaderResponse(w, http.StatusBadRequest, "Duplicate error!! Data for "+name+" already exists.")
			return
		}

		// Send the successful response back to the client
		WriteHttpHeaderResponse(w, http.StatusOK, "Successfully adding new data for "+name)
		return
	}
}

func UpdateUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Received POST request for update data")
		var BadRequest bool = false

		name := r.FormValue("name")
		nickname := r.FormValue("nickname")
		research := r.FormValue("research")

		// Update data to tbName table
		_, err := db.Exec("UPDATE tb_Name set cNickname=$1 WHERE cName=$2", nickname, name)
		if err != nil {
			fmt.Println(err)
			BadRequest = true
		} else {
			// Update data to tbResearch table
			_, err = db.Exec("UPDATE tb_Research set cResearch=$1 WHERE cName=$2", research, name)
			if err != nil {
				fmt.Println(err)
				BadRequest = true
			}
		}

		// Error from updating to database
		if BadRequest {
			WriteHttpHeaderResponse(w, http.StatusBadRequest, "Database error!! Editing data for "+name+" failed.")
			return
		}

		// Send the successful response back to the client
		WriteHttpHeaderResponse(w, http.StatusOK, "Data for "+name+" had been updated succesfully.")
		return
	}
}
func DelUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var BadRequest bool = false
		name := r.FormValue("name")
		fmt.Println("Received POST request for delete data for " + name)

		// DELETE data from tbName table
		_, err := db.Exec("DELETE FROM tb_Name WHERE cName=$1", name)
		if err != nil {
			fmt.Println(err)
			BadRequest = true
		} else {

			// DELETE data fromtbResearch table
			_, err = db.Exec("DELETE FROM tb_Research WHERE cName=$1", name)
			if err != nil {
				fmt.Println(err)
				BadRequest = true
			}
		}

		// Error from inserting to database
		if BadRequest {
			WriteHttpHeaderResponse(w, http.StatusBadRequest, "Delete data for "+name+" already failed !!")
			return
		}

		// Send the successful response back to the client
		WriteHttpHeaderResponse(w, http.StatusOK, "Successfully deleteing data for "+name)
		return
	}
}
func ShowInsertForm(w http.ResponseWriter, r *http.Request) {
	data := addeditform_data{
		Title:              "Insert student research information",
		FormAction:         "/InsertUserData",
		InputNameValue:     "",
		InputNameAttr:      "autofocus",
		InputNicknameValue: "",
		InputResearchValue: "",
		DisplayDelButton:   "none",
	}
	tmpl := template.Must(template.ParseFiles("addeditdata.html"))
	tmpl.Execute(w, data)
}
func ShowAllUserData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT tb_Name.cName, tb_Name.cNickname, tb_Research.cResearch FROM tb_Name Left Join tb_Research on tb_Research.cName=tb_Name.cName")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var data []AllData
	var editlink string
	item := 1
	for rows.Next() {
		var name, nickname, research string
		err := rows.Scan(&name, &nickname, &research)
		if err != nil {
			log.Fatal(err)
		}

		editlink = name
		data = append(data, AllData{Item: item, Name: name, Nickname: nickname, Research: research, EditLink: editlink})
		item++
	}

	tmpl := template.Must(template.ParseFiles("alldata.html"))
	tmpl.Execute(w, data)
}

func main() {
	// Connect to the PostgreSQL database
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err)
		//	http.Error(w, "Database connection error", http.StatusBadRequest)
	}
	defer db.Close()

	//HTTP Routes URL's paths
	http.HandleFunc("/", ShowHomePage)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	http.HandleFunc("/about", ShowAboutPage)
	http.HandleFunc("/about/", ShowAboutPage)

	http.HandleFunc("/ShowInsert", ShowInsertForm)

	http.HandleFunc("/userdata/", ShowUpdate)

	http.HandleFunc("/InsertUserData", InsertUserData)

	http.HandleFunc("/UpdateUserData", UpdateUserData)

	http.HandleFunc("/DelUserData", DelUserData)

	http.HandleFunc("/alldata", ShowAllUserData)

	//	http.HandleFunc("/404", ShowPageNotFound)

	fmt.Println("Server listening on :8000")
	http.ListenAndServe(":8000", nil)
}
