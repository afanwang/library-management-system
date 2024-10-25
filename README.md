# Library management system

# Design
1. DB Model:  Start with data models before the logic layers. I used Golang `sqlc` (https://github.com/rubenv/sql-migrate) to generate the postgreSQL code.
   Build an interface for the Model logic, so the DB layer can easily mocked by `pgxmock` unit-test.

2. HTTP Handlers: Write http handler code. The handlers will use the model interface. This package uses a high performance HTTP router(https://github.com/julienschmidt/httprouter) which has better performance than the obsolete Gorilla Mux(https://github.com/gorilla/mux).



| **Section**         | **Feature**               | **Status**    |
|---------------------|---------------------------|---------------|
| **Model**          		 	
|                     | User with roles            | ❌ Not Done          |		
|                     | AddUser (registration)              | ❌ Not Done         |
|                     | Book with Authors (CRUD)               | ✅ Done        |
|                     | BorrowBook                | ✅ Done    |
|                     | ReturnBook                | ✅ Done   |
| **APIs**            |       |
| | User Registration (password)        | ❌ Not Done  
|                     | Login (password)              | ❌ Not Done    |
|                     | Add New Book              | ✅ Done        |
|                     | Edit Book Details         | ✅ Done         |
|                     | Delete Book               | ✅ Done        |
|                     | View Borrowed Books       | ✅ Done     |
|                     | View All Books with Authors (better to have)      | ✅ Done     |
|                     | Borrow Book Endpoint      | ✅ Done     |
|                     | Return Book Endpoint      | ✅ Done     |
| **RBAC to APIs** | |  |
|                     | Only Admin can Delete Book     | ❌ Not Done    |
|                     | Only Admin can Edit Book     | ❌ Not Done    |
| **Documentation** |Documentation of Setup and Test     | ❌ Not Done        |
| **Test**  | Write Simple Test       |❌ Not Done        |
| **Bonus Features**  
|                     | API Rate Limiting         | ❌ Not Done    |
|                     | User Registration/Ligin using Web3              | ❌ Not Done    |
|                     | Unit Tests test                | ❌ Not Done    |


## Manual Tests


## Unit tests

