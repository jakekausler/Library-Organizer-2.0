package main

type Bookcase struct {
	ID                int64       `json:"id"`
	SpacerHeight      int64       `json:"spacerheight"`
	PaddingLeft       int64       `json:"paddingleft"`
	PaddingRight      int64       `json:"paddingright"`
	BookMargin        int64       `json:"bookmargin"`
	Width             int64       `json:"width"`
	Shelves           []Bookshelf `json:"shelves"`
	AverageBookHeight float64     `json:"averagebookheight"`
	AverageBookWidth  float64     `json:"averagebookwidth"`
}

type Bookshelf struct {
	ID     int64  `json:"id"`
	Height int64  `json:"height"`
	Books  []Book `json:"books"`
}

type StatChart struct {
	Chart StatChartInfo `json:"chart"`
	Data  []StatData    `json:"data"`
}

type StatChartInfo struct {
	Caption           string `json:"caption"`
	FormatNumberScale string `json:"formatNumberScale"`
	NumberSuffix      string `json:"numberSuffix"`
	Decimals          string `json:"decimals"`
}

type StatData struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	ToolText string `json:"tooltext"`
}

type BookIds struct {
	BookIds []string `json:"bookids"`
}

//BookSet is a collection of books, along with the number of pages of data there are without a limit imposed
type BookSet struct {
	Books         []Book `json:"books"`
	NumberOfBooks int64  `json:"numbooks"`
}

//Book is a book
type Book struct {
	ID                  string        `json:"bookid"`
	Title               string        `json:"title"`
	Subtitle            string        `json:"subtitle"`
	OriginallyPublished string        `json:"originallypublished"`
	Publisher           Publisher     `json:"publisher"`
	IsRead              bool          `json:"isread"`
	IsReference         bool          `json:"isreference"`
	IsOwned             bool          `json:"isowned"`
	ISBN                string        `json:"isbn"`
	Loanee              Name          `json:"loanee"`
	Dewey               string        `json:"dewey"`
	Pages               int64         `json:"pages"`
	Width               int64         `json:"width"`
	Height              int64         `json:"height"`
	Depth               int64         `json:"depth"`
	Weight              float64       `json:"weight"`
	PrimaryLanguage     string        `json:"primarylanguage"`
	SecondaryLanguage   string        `json:"secondarylanguage"`
	OriginalLanguage    string        `json:"originallanguage"`
	Series              string        `json:"series"`
	Volume              float64       `json:"volume"`
	Format              string        `json:"format"`
	Edition             int64         `json:"edition"`
	IsReading           bool          `json:"isreading"`
	IsShipping          bool          `json:"isshipping"`
	ImageURL            string        `json:"imageurl"`
	SpineColor          string        `json:"spinecolor"`
	CheapestNew         float64       `json:"cheapestnew"`
	CheapestUsed        float64       `json:"cheapestused"`
	EditionPublished    string        `json:"editionpublished"`
	Contributors        []Contributor `json:"contributors"`
}

//Publisher is a publisher
type Publisher struct {
	ID            string `json:"id"`
	Publisher     string `json:"publisher"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	ParentCompany string `json:"parentcompany"`
}

//Name is a name
type Name struct {
	First   string `json:"first"`
	Middles string `json:"middles"`
	Last    string `json:"last"`
}

//Contributor is a contributor
type Contributor struct {
	ID   string `json:"id"`
	Name Name   `json:"name"`
	Role string `json:"role"`
}
