# Part 1
## Library management system

See the detailed list of features and their progress.

## Design
1. DB Model:  Start with data models before the logic layers. I used Golang `sqlc` (https://github.com/rubenv/sql-migrate) to generate the postgreSQL code.
   Build an interface for the Model logic, so the DB layer can easily mocked by `pgxmock` unit-test.

2. HTTP Handlers: Write http handler code. The handlers will use the model interface. This package uses a high performance HTTP router(https://github.com/julienschmidt/httprouter) which has better performance than the obsolete Gorilla Mux(https://github.com/gorilla/mux).

3. Identity: user can be a user or an admin. The signup and login will be token-based authentication.

4. We need to add a RBAC features for the APIs. I do have experience with Casbin which may be an overkill for this task, so I choose to simply add a role-based restriction to the HTTP middleware for the authority check.

5. Some utility function:
- Added a Makefile for repeated commands.
- Addd a dockerfile for the service.

6. Use Web3 for registration and login. I am thinking of using Ethereum or Web3 library for signature verification. [Reference](https://www.dock.io/post/web3-authentication)

| **Module**         | **Feature**               | **Status**    |
|---------------------|---------------------------|---------------|
| **Model**          		 	
|                     | User with roles            |  ✅ Done           |		
|                     | AddUser (registration)              |  ✅ Done          |
|                     | Book with Authors (CRUD)               | ✅ Done        |
|                     | BorrowBook                | ✅ Done    |
|                     | ReturnBook                | ✅ Done   |
| **APIs**            |       |
|                     | User Registration (password)        |  ✅ Done   
|                     | Login (password)              |  ✅ Done    |
|                     | Add New Book              | ✅ Done        |
|                     | Edit Book Details         | ✅ Done         |
|                     | Delete Book               | ✅ Done        |
|                     | View Borrowed Books       | ✅ Done     |
|                     | View All Books with Authors (better to have)      | ✅ Done     |
|                     | Borrow Book Endpoint      | ✅ Done     |
|                     | Return Book Endpoint      | ✅ Done     |
| **RBAC to APIs** | |  |
|                     | Only Admin can Delete Book     |  ✅ Done    |
|                     | Only Admin can Edit Book     |  ✅ Done     |
| **Documentation** |Documentation of Setup and Test     | ❌ Not Done        |
| **Test**  | Write Simple Test       |❌ Not Done        |
| **Bonus Features**  
|                     | API Rate Limiting         | ❌ Not Done    |
|                     | User Registration/Ligin using Web3              | ❌ Not Done    |
|                     | Unit Tests test                | ❌ Not Done    |


### Manual Tests

### Unit tests


# Part 2
| **Module**         | **Feature**               | **Status**    |
|---------------------|---------------------------|---------------|
|                     |                 | ❌ Not Started    |
