package libraries

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	"./../books"
	"./../information"
	"./../users"
)

const (
	getLibrariesQuery = "SELECT libraries.id, name, permission, usr FROM libraries JOIN permissions ON libraries.id=permissions.libraryid join library_members on libraries.ownerid=library_members.id WHERE permissions.permission & 1 and permissions.userid=(SELECT id from library_members join usersession on library_members.id=usersession.userid WHERE sessionkey=?)"
	getBreaks         = "SELECT BreakType, ValueType, Value FROM breaks WHERE libraryid=?"
)

var logger = log.New(os.Stderr, "log: ", log.LstdFlags|log.Lshortfile)

//Break is a case break
type Break struct {
	LibraryID string    `json:"libaryid"`
	ValueType string `json:"valuetype"`
	Value     string `json:"value"`
	BreakType string `json:"breaktype"`
}

//EditedCases is a set of edited cases and ids to remove
type EditedCases struct {
	EditedCases []EditedCase `json:"editedcases"`
	ToRemoveIds []int64      `json:"toremoveids"`
	LibraryID   int64        `json:"libraryid"`
}

//EditedCase is an edited case
type EditedCase struct {
	ID              int64 `json:"id"`
	SpacerHeight    int64 `json:"spacerheight"`
	PaddingLeft     int64 `json:"paddingleft"`
	PaddingRight    int64 `json:"paddingright"`
	Width           int64 `json:"width"`
	ShelfHeight     int64 `json:"shelfheight"`
	NumberOfShelves int64 `json:"numberofshelves"`
	CaseNumber      int64 `json:"casenumber"`
}

//Bookcase is a bookcase
type Bookcase struct {
	ID                int64         `json:"id"`
	SpacerHeight      int64         `json:"spacerheight"`
	PaddingLeft       int64         `json:"paddingleft"`
	PaddingRight      int64         `json:"paddingright"`
	BookMargin        int64         `json:"bookmargin"`
	Width             int64         `json:"width"`
	Shelves           []Bookshelf   `json:"shelves"`
	AverageBookHeight float64       `json:"averagebookheight"`
	AverageBookWidth  float64       `json:"averagebookwidth"`
	Library           books.Library `json:"library"`
	CaseNumber        int64         `json:"casenumber"`
}

//Bookshelf is a shelf on a bookcase
type Bookshelf struct {
	ID     int64        `json:"id"`
	Height int64        `json:"height"`
	Books  []books.Book `json:"books"`
}

//OwnedLibrary is a library owned by the user and the users that have permission to view them
type OwnedLibrary struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Users []UserWithPermission `json:"user"`
}

//UserWithPermission is a user and permission
type UserWithPermission struct {
	ID int64 `json:"id"`
	Username string `json:"username"`
	FirstName string `json:"firstname"`
	LastName string `json:"lastname"`
	FullName string `json:"fullname"`
	Email string `json:"email"`
	IconURL string `json:"iconurl"`
	Permission int64 `json:"permission"`
}

//GetCases gets cases
func GetCases(db *sql.DB, libraryid, session string) ([]Bookcase, error) {
	authorseries, err := GetAuthorBasedSeries(db, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	sortMethod, err := GetLibrarySort(db, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	books, _, err := books.GetBooks(db, sortMethod, "both", "both", "yes", "both", "both", "both", "", "1", "-1", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", libraryid, "", session, authorseries)
	breaks, err := GetLibraryBreaks(db, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	query := "SELECT CaseId, Width, SpacerHeight, PaddingLeft, PaddingRight, BookMargin, CaseNumber, NumberOfShelves, ShelfHeight FROM bookcases WHERE libraryid=? ORDER BY CaseNumber"
	rows, err := db.Query(query, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	dim, err := information.GetDimensions(db, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	var cases []Bookcase
	for rows.Next() {
		var id, width, spacerHeight, paddingLeft, paddingRight, bookMargin, caseNumber, numberOfShelves, shelfHeight int64
		err = rows.Scan(&id, &width, &spacerHeight, &paddingLeft, &paddingRight, &bookMargin, &caseNumber, &numberOfShelves, &shelfHeight)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		avgWidth := 1;
		avgHeight := 26;
		if (dim["averagewidth"] > 0) {
			avgWidth = int(dim["averagewidth"]);
		}
		//Leave room for text
		if dim["averageheight"] > 25 {
			avgHeight = int(dim["averageheight"]);
		}
		bookcase := Bookcase{
			ID:                id,
			Width:             width,
			SpacerHeight:      spacerHeight,
			PaddingLeft:       paddingLeft,
			PaddingRight:      paddingRight,
			BookMargin:        bookMargin,
			AverageBookWidth:  float64(avgWidth),
			AverageBookHeight: float64(avgHeight),
			CaseNumber:        caseNumber,
		}
		// 		shelfquery := "SELECT id, Height FROM shelves WHERE CaseId=? ORDER BY ShelfNumber"
		// 		shelfrows, err := db.Query(shelfquery, id)
		// 		if err != nil {
		// logger.Printf("Error: %+v", err)
		// return nil, err
		// 		}
		// 		for shelfrows.Next() {
		// 			var shelfid, height int64
		// 			shelfrows.Scan(&shelfid, &height)
		// 			bookcase.Shelves = append(bookcase.Shelves, Bookshelf{
		// 				ID:     shelfid,
		// 				Height: height,
		// 			})
		// 		}
		for i := 0; i < int(numberOfShelves); i++ {
			bookcase.Shelves = append(bookcase.Shelves, Bookshelf{
				ID:     0,
				Height: shelfHeight,
			})
		}
		cases = append(cases, bookcase)
	}
	index := 0
	x := 0
	for c, bookcase := range cases {
		breakcase := false
		for s := range bookcase.Shelves {
			if breakcase {
				break
			}
			x = int(bookcase.PaddingLeft)
			useWidth := int(bookcase.AverageBookWidth)
			if index < len(books) && books[index].Width > 0 {
				useWidth = int(books[index].Width)
			}
			breakshelf := false
			for index < len(books) && useWidth+x <= int(bookcase.Width)-int(bookcase.PaddingRight) {
				if breakshelf || breakcase {
					break
				}
				cases[c].Shelves[s].Books = append(cases[c].Shelves[s].Books, books[index])
				x += useWidth
				index++
				useWidth = int(bookcase.AverageBookWidth)
				if index < len(books) && books[index].Width != 0 {
					useWidth = int(books[index].Width)
				}
				breakInner := false
				for _, b := range breaks {
					if breakInner {
						break
					}
					switch b.ValueType {
					case "ID":
						if b.Value == books[index].ID {
							breakInner = true;
							if b.BreakType == "CASE" {
								breakcase = true
							} else {
								breakshelf = true
							}
							break
						}
					case "DEWEY":
						if index != 0 && books[index-1].Dewey.Valid && books[index-1].Dewey.String < b.Value && books[index].Dewey.Valid && books[index].Dewey.String >= b.Value {
							breakInner = true
							if b.BreakType == "CASE" {
								breakcase = true
							} else {
								breakshelf = true
							}
							break
						}
					case "SERIES":
						if index != 0 && books[index-1].Series < b.Value && books[index].Series >= b.Value {
							breakInner = true
							if b.BreakType == "CASE" {
								breakcase = true
							} else {
								breakshelf = true
							}
							break
						}
					}
				}
			}
		}
	}
	return cases, nil
}

//GetLibraries gets the libraries available to a user
func GetLibraries(db *sql.DB, session string) ([]books.Library, error) {
	var libraries []books.Library
	rows, err := db.Query(getLibrariesQuery, session)
	if err != nil {
		logger.Printf("%+v", err)
		return nil, err
	}
	for rows.Next() {
		var l books.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Permissions, &l.Owner); err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		libraries = append(libraries, l)
	}
	return libraries, nil
}

//SaveCases saves book cases
func SaveCases(db *sql.DB, libraryid string, cases EditedCases) error {
	for _, c := range cases.EditedCases {
		id := c.ID
		caseNumber := c.CaseNumber
		width := c.Width
		spacerHeight := c.SpacerHeight
		paddingLeft := c.PaddingLeft
		paddingRight := c.PaddingRight
		shelfHeight := c.ShelfHeight
		numberOfShelves := c.NumberOfShelves
		if id == 0 {
			query := "INSERT INTO bookcases (casenumber, width, spacerheight, paddingleft, paddingright, libraryid, numberofshelves, shelfheight) VALUES (?,?,?,?,?,?,?,?)"
			_, err := db.Exec(query, caseNumber, width, spacerHeight, paddingLeft, paddingRight, libraryid, numberOfShelves, shelfHeight)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		} else {
			query := "UPDATE bookcases SET casenumber=?, width=?, spacerheight=?, paddingleft=?, paddingright=?, libraryid=?, numberOfShelves=?, shelfheight=? WHERE caseid=?"
			_, err := db.Exec(query, caseNumber, width, spacerHeight, paddingLeft, paddingRight, libraryid, numberOfShelves, shelfHeight, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	var removedCases []string
	for _, id := range cases.ToRemoveIds {
		i := strconv.FormatInt(id, 10)
		removedCases = append(removedCases, i)
	}
	if len(removedCases) > 0 {
		query := "DELETE FROM bookcases WHERE CaseId IN (" + strings.Join(removedCases, ",") + ")"
		_, err := db.Exec(query)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
	}
	return nil
}

//AddBreak adds a shelf break
func UpdateBreaks(db *sql.DB, libraryid string, breaks []Break) error {
	query := "DELETE FROM breaks WHERE libraryid=?"
	_, err := db.Exec(query, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	for _, b := range breaks {
		query = "INSERT INTO breaks (libraryid, breaktype, valuetype, value) VALUES (?,?,?,?)"
		_, err = db.Exec(query, libraryid, b.BreakType, b.ValueType, b.Value)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
	}
	return nil
}

//GetLibraryBreaks gets shelf breaks
func GetLibraryBreaks(db *sql.DB, libraryid string) ([]Break, error) {
	var breaks []Break
	query := "SELECT breaktype, valuetype, value FROM breaks WHERE libraryid=?"
	rows, err := db.Query(query, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	for rows.Next() {
		var b Break
		b.LibraryID = libraryid
		err = rows.Scan(&b.BreakType, &b.ValueType, &b.Value)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		breaks = append(breaks, b)
	}
	return breaks, nil
}

//GetOwnedLibraries gets owned libraries and the people who have permission to do things with them
func GetOwnedLibraries(db *sql.DB, session string) ([]OwnedLibrary, error) {
	userid, err := users.GetUserID(db, session)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	var libraries []OwnedLibrary
	query := "SELECT libraries.id, libraries.name FROM libraries WHERE libraries.ownerid=?"
	rows, err := db.Query(query, userid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	for rows.Next() {
		var library OwnedLibrary
		err = rows.Scan(&library.ID, &library.Name)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		query = "SELECT library_members.id, library_members.usr, library_members.firstname, library_members.lastname, library_members.email, library_members.iconurl, permissions.permission FROM libraries JOIN permissions ON libraries.id=permissions.libraryid JOIN library_members ON permissions.userid=library_members.id WHERE libraries.id=? AND permissions.userid != ?"
		innerRows, err := db.Query(query, library.ID, userid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		for innerRows.Next() {
			var user UserWithPermission
			err := innerRows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.IconURL, &user.Permission)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return nil, err
			}
			user.FullName = user.FirstName + " " + user.LastName
			library.Users = append(library.Users, user)
		}
		libraries = append(libraries, library)
	}
	return libraries, nil
}

//SaveOwnedLibraries saves owned libraries and the people who have permission to do things with them
func SaveOwnedLibraries(db *sql.DB, ownedLibraries []OwnedLibrary, session string) error {
	userid, err := users.GetUserID(db, session)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	query := "DELETE FROM libraries WHERE ownerid=?"
	_, err = db.Exec(query, userid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	for _, ownedLibrary := range ownedLibraries {
		oldID := ownedLibrary.ID
		query = "INSERT INTO libraries (name, ownerid, sortmethod) VALUES (?,?)"
		res, err := db.Exec(query, ownedLibrary.Name, userid, information.SORTMETHOD)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		ownedLibrary.ID, err = res.LastInsertId()
		query = "DELETE FROM permissions WHERE libraryid=?"
		_, err = db.Exec(query, ownedLibrary.ID)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		query = "DELETE FROM permissions WHERE libraryid=?"
		_, err = db.Exec(query, oldID)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		if oldID != -1 {
			query = "UPDATE books SET libraryid=? WHERE libraryid=?"
			_, err = db.Exec(query, ownedLibrary.ID, oldID)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
			query = "UPDATE bookcases SET libraryid=? WHERE libraryid=?"
			_, err = db.Exec(query, ownedLibrary.ID, oldID)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
			query = "UPDATE breaks SET libraryid=? WHERE libraryid=?"
			_, err = db.Exec(query, ownedLibrary.ID, oldID)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
			query = "UPDATE series_author_sorts SET libraryid=? WHERE libraryid=?"
			_, err = db.Exec(query, ownedLibrary.ID, oldID)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
		query = "INSERT INTO permissions (userid, libraryid, permission) VALUES (?,?,7)"
		_, err = db.Exec(query, userid, ownedLibrary.ID)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		for _, user := range ownedLibrary.Users {
			query = "INSERT INTO permissions (userid, libraryid, permission) VALUES (?,?,?)"
			_, err = db.Exec(query, user.ID, ownedLibrary.ID, user.Permission)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	query = "SELECT DISTINCT(libraryid) FROM books";
	rows, err := db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var booklibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		booklibraryids = append(booklibraryids, lid)
	}
	query = "SELECT DISTINCT(libraryid) FROM permissions";
	rows, err = db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var permissionslibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		permissionslibraryids = append(permissionslibraryids, lid)
	}
	query = "SELECT DISTINCT(libraryid) FROM bookcases";
	rows, err = db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var bookcaseslibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		bookcaseslibraryids = append(bookcaseslibraryids, lid)
	}
	query = "SELECT DISTINCT(libraryid) FROM breaks";
	rows, err = db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var breakslibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		breakslibraryids = append(breakslibraryids, lid)
	}
	query = "SELECT DISTINCT(libraryid) FROM series_author_sorts";
	rows, err = db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var serieslibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		serieslibraryids = append(serieslibraryids, lid)
	}
	query = "SELECT id FROM libraries"
	rows, err = db.Query(query)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	var actuallibraryids []int64
	for rows.Next() {
		var lid int64
		err := rows.Scan(&lid)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
		actuallibraryids = append(actuallibraryids, lid)
	}
	for _, id := range booklibraryids {
		if !contains(actuallibraryids, id) {
			query = "DELETE FROM books WHERE libraryid=?"
			_, err := db.Exec(query, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	for _, id := range permissionslibraryids {
		if !contains(actuallibraryids, id) {
			query = "DELETE FROM books WHERE libraryid=?"
			_, err := db.Exec(query, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	for _, id := range breakslibraryids {
		if !contains(actuallibraryids, id) {
			query = "DELETE FROM books WHERE libraryid=?"
			_, err := db.Exec(query, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	for _, id := range bookcaseslibraryids {
		if !contains(actuallibraryids, id) {
			query = "DELETE FROM books WHERE libraryid=?"
			_, err := db.Exec(query, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	for _, id := range serieslibraryids {
		if !contains(actuallibraryids, id) {
			query = "DELETE FROM books WHERE libraryid=?"
			_, err := db.Exec(query, id)
			if err != nil {
				logger.Printf("Error: %+v", err)
				return err
			}
		}
	}
	return nil
}

func contains(s []int64, e int64) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

//GetAuthorBasedSeries gets series that are sorted by author then title, instead of volume
func GetAuthorBasedSeries(db *sql.DB, libraryid string) ([]string, error) {
	var series []string
	query := "SELECT series FROM series_author_sorts WHERE libraryid=?"
	rows, err := db.Query(query, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return nil, err
	}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}



//UpdateAuthorBasedSeries updates series that are sorted by author then title, instead of volume
func UpdateAuthorBasedSeries(db *sql.DB, libraryid string, series []string) error {
	query := "DELETE FROM series_author_sorts WHERE libraryid=?"
	_, err := db.Exec(query, libraryid)
	if err != nil {
		logger.Printf("Error: %+v", err)
		return err
	}
	for _, s := range series {
		query = "INSERT INTO series_author_sorts (libraryid, series) VALUES (?,?)"
		_, err = db.Exec(query, libraryid, s)
		if err != nil {
			logger.Printf("Error: %+v", err)
			return err
		}
	}
	return nil
}

//GetLibrarySort gets the sort method of a library
func GetLibrarySort(db *sql.DB, libraryid string) (string, error) {
	var method string
	query := "SELECT sortmethod FROM libraries WHERE id=?"
	err := db.QueryRow(query, libraryid).Scan(&method)
	return method, err
}



//UpdateLibrarySort updates the sort method of a library
func UpdateLibrarySort(db *sql.DB, libraryid string, method string) error {
	query := "UPDATE libraries SET sortmethod=? WHERE id=?"
	_, err := db.Exec(query, method, libraryid)
	return err
}