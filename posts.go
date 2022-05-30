package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type postDisplay struct {
	PostID        string
	Username      string
	PostCategory  string
	Likes         int
	Dislikes      int
	TitleText     string
	PostText      string
	Image         string
	ImgExists     bool
	CookieChecker bool
	Comments      []commentStruct
}

type commentStruct struct {
	CommentID       string
	CpostID         string
	CommentUsername string
	CommentText     string
	Likes           int
	Dislikes        int
	CookieChecker   bool
}

//newPost creates a new post by a registered user
func newPost(userName, category, title, post string, imageName string, db *sql.DB) {
	//If the title is empty the form is resubitting once all values have been reset so the post shouldn't be added to the database
	if title == "" {
		return
	}

	fmt.Println("ADDING POST")
	uuid := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO posts (postID, userName, category, likes, dislikes, title, post, image) VALUES (?, ?, ?, 0, 0, ?, ?, ?)", uuid, userName, category, title, post, imageName)
	if err != nil {
		fmt.Println("Error adding new post")
		log.Fatal(err.Error())
	}
	Person.PostAdded = true

	//Add the post to the Category table with relevant categories selected
	//Split the category string by Spaces to see which categories are selected
	catSlc := strings.Split(category, " ")
	feSelected := 0
	beSelected := 0
	fsSelected := 0
	//Loop through categories if any element = Frontend backend or fullstack chane accordingly
	for _, r := range catSlc {
		if r == "FrontEnd" {
			feSelected = 1
		} else if r == "BackEnd" {
			beSelected = 1
		} else if r == "FullStack" {
			fsSelected = 1
		}
	}

	//Now add the relevant value to the category in the category table
	_, errAddCats := db.Exec("INSERT INTO categories (postID, FrontEnd, BackEnd, FullStack) VALUES (?, ?, ?, ?)", uuid, feSelected, beSelected, fsSelected)
	if errAddCats != nil {
		fmt.Println("ERROR when adding into the category table")
	}
}

func postData(db *sql.DB) []postDisplay {
	// rows, err := db.Query("SELECT postID, userName, category, likes, dislikes, title, post FROM posts")
	rows, err := db.Query("SELECT postID, userName, category, likes, dislikes, title, post, image FROM posts")

	if err != nil {
		fmt.Println("Error selecting post data")
		log.Fatal(err.Error())
	}

	finalArray := []postDisplay{}

	for rows.Next() {

		var u postDisplay
		err := rows.Scan(
			&u.PostID,
			&u.Username,
			&u.PostCategory,
			&u.Likes,
			&u.Dislikes,
			&u.TitleText,
			&u.PostText,
			&u.Image,
		)
		u.CookieChecker = Person.CookieChecker
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}

		if u.Image != "" {
			u.ImgExists = true
		}


		//Get comments for the relevent post]
		//Make []commentstruct to hold all the comments
		commentSlc := []commentStruct{}
		var tempComStruct commentStruct
		fmt.Println("-------------------------------This line is the post ID: " + u.PostID)

		commentRow, errComs := db.Query("SELECT commentID, postID, username, commentText, likes, dislikes FROM comments WHERE postID = ?", u.PostID)
		if errComs != nil {
			fmt.Println("Error selecting comment data")
			log.Fatal(errComs.Error())
		}
		for commentRow.Next() {
			err := commentRow.Scan(
				&tempComStruct.CommentID,
				&tempComStruct.CpostID,
				&tempComStruct.CommentUsername,
				&tempComStruct.CommentText,
				&tempComStruct.Likes,
				&tempComStruct.Dislikes,
			)
			tempComStruct.CookieChecker = Person.CookieChecker
			if err != nil {
				fmt.Println("Error scanning comments")
				log.Fatal(err.Error())
			}
			fmt.Printf("\nCOMMENT STRUCT_____-------------------------------------%v\n\n", tempComStruct)
			commentSlc = append(commentSlc, tempComStruct)
		}
		u.Comments = commentSlc

		//Append all post information to the finalarray
		finalArray = append(finalArray, u)

		//Reverse finalArray so that the most recently added posts show first
		for i, j := 0, len(finalArray)-1; i < j; i, j = i+1, j-1 {
			finalArray[i], finalArray[j] = finalArray[j], finalArray[i]
		}
	}
	return finalArray
}

func LikeButton(postID string, db *sql.DB) {
	//Check if the user has already liked this post/comment
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE postID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err2 := findRow.Scan(
			&check.Likes,
		)

		if err2 != nil {
			log.Fatal(err2.Error())
		}
	}

	//If rounds still equals 0 no row was found so we can insert the relevant row into our liketable
	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, postID, reference) VALUES (?, ?, 1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (LIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}

		//Increase likes
		LikeIncrease(postID, sqliteDatabase)
	} else {
		//Reference is equal to 1 so we need to undo the like action
		if check.Likes == 1 {
			LikeUndo(postID, sqliteDatabase)
			//Update reference to 0
			RefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == -1 {
			//user has already disliked so we must undislike the post and set it as liked
			DislikeUndo(postID, sqliteDatabase)
			LikeIncrease(postID, sqliteDatabase)
			//Update reference equal to 1

			RefUpdate(1, postID, sqliteDatabase)

		} else if check.Likes == 0 {
			//Increase likes only
			LikeIncrease(postID, sqliteDatabase)
			//set reference to 1
			RefUpdate(1, postID, sqliteDatabase)

		}
	}
}

func RefUpdate(value int, postID string, db *sql.DB) {
	_, err2 := db.Exec("UPDATE liketable SET reference = (?) WHERE postID = (?) AND user = (?)", value, postID, Person.Username)
	if err2 != nil {
		fmt.Println("UPDATING REFERENCE ")
		log.Fatal(err2.Error())
	}
}
func CommentRefUpdate(value int, postID string, db *sql.DB) {
	_, err2 := db.Exec("UPDATE liketable SET reference = (?) WHERE commentID = (?) AND user = (?)", value, postID, Person.Username)
	if err2 != nil {
		fmt.Println("UPDATING REFERENCE ")
		log.Fatal(err2.Error())
	}
}

func LikeIncrease(postID string, db *sql.DB) {
	//Increase likes

	likes, err := db.Query("SELECT likes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error selecting likes")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes++
	_, err2 := db.Exec("UPDATE posts SET likes = (?) WHERE postID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("UPDATING LIKES WHEN ROUNDS == 0")
		log.Fatal(err.Error())
	}

}

func LikeUndo(postID string, db *sql.DB) {
	likes, err := db.Query("SELECT likes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in LIKE UNDO")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes--
	_, err2 := db.Exec("UPDATE posts SET likes = (?) WHERE postID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("LIKE UNDO")
		log.Fatal(err.Error())
	}
}

func DislikeIncrease(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeIncrease")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes++
	_, err2 := db.Exec("UPDATE posts SET dislikes = (?) WHERE postID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("UPDATING DISLIKES")
		log.Fatal(err.Error())
	}
}

func DislikeUndo(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM posts WHERE postID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeUndo")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes--
	_, err2 := db.Exec("UPDATE posts SET dislikes = (?) WHERE postID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("DISLIKE UNDO")
		log.Fatal(err.Error())
	}
}

func DislikeButton(postID string, db *sql.DB) {
	//Check if the user has already liked/disliked this post/comment
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE postID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err := findRow.Scan(
			&check.Likes,
		)

		if err != nil {
			fmt.Println("Error in Dislike Button")
			log.Fatal(err.Error())
		}
	}

	//if rounds == 0 the user hasnt liked or disliked this post/comment yet
	if rounds == 0 {
		//Add the user to the liketable
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, postID, reference) VALUES (?, ?, -1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (DISLIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}
		//Increase number of dslikes
		DislikeIncrease(postID, sqliteDatabase)

	} else {
		if check.Likes == -1 {
			//The user has already disliked so we need to undo the dislike action
			DislikeUndo(postID, sqliteDatabase)
			//Change reference to 0
			RefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == 1 {
			//User has previously liked so we need to undo the like and dislike the comment
			//Undo like
			LikeUndo(postID, sqliteDatabase)
			// Increase dislike
			DislikeIncrease(postID, sqliteDatabase)
			//Set reference equal to -1
			RefUpdate(-1, postID, sqliteDatabase)
		} else if check.Likes == 0 {
			//The user is not currently liking or disliking the post so we need to increase dislike
			DislikeIncrease(postID, sqliteDatabase)
			//update reference to -1
			RefUpdate(-1, postID, sqliteDatabase)
		}
	}
}

//Add a new comment to a post
func newComment(userName, postID, commentText string, db *sql.DB) {
	if commentText == "" {
		return
	}

	fmt.Println("ADDING Comment")
	uuid := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO comments (commentID, postID, userName, commentText, likes, dislikes) VALUES (?, ?, ?, ?, 0, 0)", uuid, postID, userName, commentText)
	if err != nil {
		fmt.Println("ERROR ADDING COMMENT TO THE TABLE")
		log.Fatal(err.Error())
	}
	Person.PostAdded = true

}

func CommentLikeButton(postID string, db *sql.DB) {
	//Check if the user has already liked this post/comment
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE commentID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err2 := findRow.Scan(
			&check.Likes,
		)

		if err2 != nil {
			log.Fatal(err2.Error())
		}
	}

	//If rounds still equals 0 no row was found so we can insert the relevant row into our liketable
	if rounds == 0 {
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, commentID, reference) VALUES (?, ?, 1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (LIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}

		//Increase likes
		CommentLikeIncrease(postID, sqliteDatabase)
	} else {
		//Reference is equal to 1 so we need to undo the like action

		fmt.Printf("\n ------------------------------------------------------------------REFERENCE is equal to: %v", check.Likes)
		if check.Likes == 1 {
			CommentLikeUndo(postID, sqliteDatabase)
			//Update reference to 0
			CommentRefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == -1 {
			//user has already disliked so we must undislike the post and set it as liked
			CommentDislikeUndo(postID, sqliteDatabase)
			CommentLikeIncrease(postID, sqliteDatabase)
			//Update reference equal to 1

			CommentRefUpdate(1, postID, sqliteDatabase)

		} else if check.Likes == 0 {
			//Increase likes only
			CommentLikeIncrease(postID, sqliteDatabase)
			//set reference to 1
			CommentRefUpdate(1, postID, sqliteDatabase)

		}
	}
}

func CommentLikeIncrease(postID string, db *sql.DB) {
	//Increase likes

	likes, err := db.Query("SELECT likes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error selecting likes")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}
	fmt.Printf("CURRENT COMMENT LIKES: %v \n", temp.Likes)

	temp.Likes++
	fmt.Printf("\n INCREAED COMMENT LIKES: %v !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!111\n", temp.Likes)
	_, err2 := db.Exec("UPDATE comments SET likes = (?) WHERE commentID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("UPDATING LIKES WHEN ROUNDS == 0")
		log.Fatal(err.Error())
	}

}

func CommentLikeUndo(postID string, db *sql.DB) {
	likes, err := db.Query("SELECT likes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in LIKE UNDO")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for likes.Next() {
		err := likes.Scan(
			&temp.Likes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Likes--
	_, err2 := db.Exec("UPDATE comments SET likes = (?) WHERE commentID = (?)", temp.Likes, postID)
	if err2 != nil {
		fmt.Println("LIKE UNDO")
		log.Fatal(err.Error())
	}
}

func CommentDislikeUndo(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeUndo")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes--
	_, err2 := db.Exec("UPDATE comments SET dislikes = (?) WHERE commentID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("DISLIKE UNDO")
		log.Fatal(err.Error())
	}
}

func CommentDislikeIncrease(postID string, db *sql.DB) {
	dislikes, err := db.Query("SELECT dislikes FROM comments WHERE commentID = (?)", postID)
	if err != nil {
		fmt.Println("Error in DislikeIncrease")
		log.Fatal(err.Error())
	}

	var temp postDisplay
	for dislikes.Next() {
		err := dislikes.Scan(
			&temp.Dislikes,
		)
		if err != nil {
			fmt.Println("SCANNING ERROR")
			log.Fatal(err.Error())
		}
	}

	temp.Dislikes++
	_, err2 := db.Exec("UPDATE comments SET dislikes = (?) WHERE commentID = (?)", temp.Dislikes, postID)
	if err2 != nil {
		fmt.Println("UPDATING DISLIKES")
		log.Fatal(err.Error())
	}
}

func CommentDislikeButton(postID string, db *sql.DB) {
	//Check if the user has already liked/disliked this post/comment
	findRow, errRows := db.Query("SELECT reference FROM liketable WHERE commentID = (?) AND user = (?)", postID, Person.Username)
	if errRows != nil {
		fmt.Println("SELECTING LIKE ERROR")
		log.Fatal(errRows.Error())
	}
	rounds := 0

	var check postDisplay
	for findRow.Next() {
		rounds++
		err := findRow.Scan(
			&check.Likes,
		)

		if err != nil {
			fmt.Println("Error in Dislike Button")
			log.Fatal(err.Error())
		}
	}

	//if rounds == 0 the user hasnt liked or disliked this post/comment yet
	if rounds == 0 {
		//Add the user to the liketable
		_, insertLikeErr := db.Exec("INSERT INTO liketable (user, commentID, reference) VALUES (?, ?, -1)", Person.Username, postID)
		if insertLikeErr != nil {
			fmt.Println("Error when inserting into like table initially (DISLIKEBUTTON)")
			log.Fatal(insertLikeErr.Error())
		}
		//Increase number of dslikes
		CommentDislikeIncrease(postID, sqliteDatabase)

	} else {
		if check.Likes == -1 {
			//The user has already disliked so we need to undo the dislike action
			CommentDislikeUndo(postID, sqliteDatabase)
			//Change reference to 0
			CommentRefUpdate(0, postID, sqliteDatabase)
		} else if check.Likes == 1 {
			//User has previously liked so we need to undo the like and dislike the comment
			//Undo like
			CommentLikeUndo(postID, sqliteDatabase)
			// Increase dislike
			CommentDislikeIncrease(postID, sqliteDatabase)
			//Set reference equal to -1
			CommentRefUpdate(-1, postID, sqliteDatabase)
		} else if check.Likes == 0 {
			//The user is not currently liking or disliking the post so we need to increase dislike
			CommentDislikeIncrease(postID, sqliteDatabase)
			//update reference to -1
			CommentRefUpdate(-1, postID, sqliteDatabase)
		}
	}
}

//Make a function that takes in a slice of postIDs and returns a slice of poststructs with all the details

func PostGetter(postIDSlc []string, db *sql.DB) []postDisplay {
	//Create a slice of postdetail structs
	finalArray := []postDisplay{}
	//loop through the slice to get each postID
	for _, r := range postIDSlc {
		rows, errDetails := db.Query("SELECT postID, userName, category, likes, dislikes, title, post FROM posts WHERE postID = (?)", r)
		if errDetails != nil {
			fmt.Println("ERROR when selecting the information for specific posts (func POSTGETTER)")
			log.Fatal(errDetails.Error())
		}

		for rows.Next() {
			var postDetails postDisplay
			err := rows.Scan(
				&postDetails.PostID,
				&postDetails.Username,
				&postDetails.PostCategory,
				&postDetails.Likes,
				&postDetails.Dislikes,
				&postDetails.TitleText,
				&postDetails.PostText,
			)
			postDetails.CookieChecker = Person.CookieChecker
			if err != nil {
				fmt.Println("ERROR Scanning through the rows (func POSTGETTER)")
				log.Fatal(err.Error())
			}
			finalArray = append(finalArray, postDetails)
		}
	}
	return finalArray
}
