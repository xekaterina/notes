package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Note struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func CreateNoteTable(db *sql.DB) error {
	// Создание таблицы "notes", если она не существует
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		title TEXT,
		content TEXT
	)`)
	return err
}

func CreateNote(c *gin.Context) {

	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var note Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подготовка запроса
	stmt, err := db.Prepare("INSERT INTO notes (title, content) VALUES ($1, $2)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	// Выполнение запроса
	_, err = stmt.Exec(note.Title, note.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Note created successfully"})

}

func NoteList(c *gin.Context) {

	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Подготовка запроса
	rows, err := db.Query("SELECT id, title, content FROM notes")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Println("Error querying database:", err)
		return
	}

	// Выполнение запроса
	notes := []Note{}
	for rows.Next() {
		var note Note
		err = rows.Scan(&note.ID, &note.Title, &note.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Println("Error scanning rows:", err)
			return
		}
		notes = append(notes, note)
	}

	// Отправка ответа в формате JSON
	c.JSON(http.StatusOK, gin.H{
		"data": notes,
	})
}

func DeleteNote(c *gin.Context) {

	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	//подготовка запроса
	stmt, err := db.Prepare("DELETE FROM notes WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	//выполнение запроса
	res, err := stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func NoteById(c *gin.Context) {
	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note Note
	err = db.QueryRow("SELECT id, title, content FROM notes WHERE id=$1", id).Scan(&note.ID, &note.Title, &note.Content)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

func UpNote(c *gin.Context) {
	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подготовка запроса
	stmt, err := db.Prepare("UPDATE notes SET title=$1, content=$2 WHERE id=$3")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	// Выполнение запроса
	res, err := stmt.Exec(note.Title, note.Content, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully"})

}

func main() {

	db, err := sql.Open("postgres", "postgres://postgres:dwq21d21@db/notes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание таблицы "notes" при запуске приложения
	if err := CreateNoteTable(db); err != nil {
		log.Fatal("Error creating note table:", err)
	}

	r := gin.Default()

	r.POST("/createnote", CreateNote)
	r.GET("/notelist", NoteList)
	r.GET("/notebyid", NoteById)
	r.PUT("/upnote", UpNote)
	r.DELETE("/deletenote", DeleteNote)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
