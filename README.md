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
1. Filter and Pagination
    ```Go
        // Define the struct
        type SampleFilter struct {
            // filters
            User_Id       *int   `json:"user_id"`
            Title         string `json:"title"`
            Title_Opt     string `json:"title_opt"` // to use custom operator other than equal
            CategoryTag   string `json:"categoryTag"`
            Status        Status `json:"status"`
            // pagination
            Page         int     `json:"page"`
            PageSize     int     `json:"pagesize"`
            // if you don't want pagination use this instead
            NoPagination bool
        }

        // Create instance of the struct
        // note that most of the time the variables is coming from query string param
        filter := SampleFilter{
            Title: title, // from variable title
            Title_Opt: "left", // produces: "Title LIKE '....%'",
            Page: page, // from variable page
            PageSize: pagesize, // from variable pagesize
            // for no pagination case
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