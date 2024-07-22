# getuk
Helper package for package GORM, mostly related to scopes.

## Features
1. Filter
    - Based on struct
    - Covers several operators
        - Equal
        - Not Equal
        - Less Than
        - Less Than or Equal
        - Greater Than
        - Greater Than or Equal
        - Between
        - In
2. Pagination
    - Based on struct fixed for the follwoing attributes:
        - Page
        - PageSize
        - NoPagination


## Usage
### Filter and Pagination
```Go
// Define the struct
type SampleFilter struct {
    // filters, dynamic fields based on the table
    User_Id       *int   `json:"user_id"`
    Title         string `json:"title"`
    Title_Opt     string `json:"title_opt"` // to use custom operator other than equal
    CategoryTag   string `json:"categoryTag"`
    Status        Status `json:"status"`

    // pagination, fixed fields
    Page         int     `json:"page"`
    PageSize     int     `json:"page_size"`
    // if you want to enable non-paginated result use this, and gives true value in the instance
    NoPagination bool
}

// Create instance of the struct
// note that most of the time the variables is coming from query string param
filter := SampleFilter{
    Title: title, // from variable title
    Title_Opt: "left" // produces: "Title LIKE '....%'",
    Page: page, // from variable page
    PageSize: 10,
    // non paginated result
    NoPagination: true,
}

// Use in Gorm, assuming DB is a gorm instance
pagination := getuk.Pagination{} // to strore pagination information
result := DB.
    Model(&m.Blog{}).
    Scopes(getuk.Filter(filter)).
    Scopes(getuk.Paginate(filter, &pagination)).
    Find(&data)
```

### Flat Join
To generate a flat field from joining one table with others. Due to Gorm's nature that doesn't allow
multiple times call of `Select()` function, a pointer of string is needed to save the selection
generated.

It uses `FlatJoinOpt` type for the configartion with the following definition:
```Go
// Type
type FlatJoinOpt struct {
	Src      string   // required, source table that has ForeignKey
	SrcFkCol string   // optional, default: [Ref]_[RefCol]
	Ref      string   // required, reference table that will be looked up to
	RefCol   string   // optional, default: Id
	Mode     Mode     // optional, joining mode, default: 
	Clause   string   // optional, if addional clause needed in the JOIN default: ""
	Prefix   string   // optional, prefix for selected reference cols, default: [Ref]_
	Cols     []string // required at least one item, the cols that will be selected from Ref
}

// Const
const JMInner Mode = "INNER JOIN"
const JMOutter Mode = "CROSS JOIN"
const JMLeft Mode = "LEFT JOIN"
const JMRight Mode = "RIGHT JOIN"
```


Example:
```Go
// prep
rc := "Code"
selectStr := "SELECT User.*"
src := "User"

// Use in Gorm, assuming DB is a gorm instance
// In the following sample, it is assumed there are table User and Province
// which can be connection through table User's foreign key.
result := DB.
    // User table
    // ...

    // Province table
    // ...

    // Flat JoinDTO
    type UserListItemDto struct{
        // From User,
        Id            uint
        Name          string
        Email         string
        Province_Code string

        // From flat join
        Province_Name string
    }

    // the DTO'sintance
    data := UserListItemDto{}

    Model(&e.User{}). // Model user
    Scopes(gh.FlatJoin(&selectStr, getuk.FlatJoinOpt{Ref: "Province", RefCol: rc, Src: src, Mode: getuk.JMLeft, Cols: []string{"Name"}})).
	Select(selectStr).
    Scan(&data);
```
