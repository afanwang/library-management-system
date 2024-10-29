# Table of Contents

- [Part 1](#part-1)
   - [Design](#design)
   - [List of Features](#list-of-features)
   - [Api Tests](#api-tests)
- [Part 2](#part-2)
   - [Version 1 Result](#version-1)
   - [Version 2 Result](#version-2)
- [Developer Setup](#setup)

# Part 1
## Library management system

See the detailed list of features and their progress.

## Design
1. DB Model:  Start with data models before the logic layers. I used Golang `sqlc` (https://github.com/rubenv/sql-migrate) to generate the postgreSQL code.
   Build an interface for the Model logic, so the DB layer can easily mocked by `pgxmock` unit-test.
   
   Database selection and reason: I choose PostgreSQL as it is currently the most used DB and it supports [mvcc](https://www.postgresql.org/docs/7.1/mvcc.html) which has much better concurrency; In addition, PostgreSQL's extensions make it easily be suitable for any usecase in the future. Read [1000+ PostgresQL extensions](https://gist.github.com/joelonsql/e5aa27f8cc9bd22b8999b7de8aee9d47). For example, the `timescaleDB` and `pgVector` are pretty popular. 

2. HTTP Handlers: Write http handler code. The handlers will use the model interface. This package uses a high performance HTTP router(https://github.com/julienschmidt/httprouter) which has better performance than the obsolete Gorilla Mux(https://github.com/gorilla/mux).

3. Identity: user can be a user or an admin. The signup and login will be token-based authentication.

4. We need to add a RBAC features for the APIs. I do have experience with Casbin which may be an overkill for this task, so I choose to simply add a role-based restriction to the HTTP middleware for the authority check.

5. Some utility function:
- Added a Makefile for repeated commands.
- Addd a dockerfile for the service.

6. Use Web3 for registration and login. I am thinking of using Ethereum or Web3 library for signature verification. [Reference](https://www.dock.io/post/web3-authentication)
- User provides their Ethereum address, server responds with a nonce tied to the user, and user signs the nonce with their Ethereum private key. Client sends the signature to the server, server verifies the signature and authenticates the user.
- Added an interface to support Password-auth and Web3 auth.

## List of Features
| **Module**         | **Feature**               | **Status**    |
|---------------------|---------------------------|---------------|
| **Model**          		 	
|                     | User with roles            |  ✅ Done           |		
|                     | AddUser (registration)              |  ✅ Done          |
|                     | Book with Authors (CRUD)               | ✅  Done        |
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
| **Documentation** |Documentation of Setup and Test     |  ✅ Done        |
| **Test**  | API Integration Tests       | ✅ Done      |
| **Bonus Features**  
|                     | API Rate Limiting         | ✅ Done      |
|                     | User Registration/Ligin using Web3              | ✅ Done     |


### Api Tests
Refer to `api-integration-test.go`


# Part 2
| **Module**         | **Feature**               | **Status**    |
|---------------------|---------------------------|---------------|
|  Version 1: sequential processing          |  Took 34 sec               | ✅ Done  |
|  Version 2: parrallel processing          |  Took 6 sec               | ✅ Done  |

## Version 1
It took 34 seconds to process
```
Total event counts across the block range:
0x804c9b842b2748a22bb64b345453a3de7ca54a6ca45ce00d415894979e22897a: 13
0x0c5bc74ccdf848b38eb526a154b85085e1d61addf1d100cba2074e039c0b6340: 4
0x02ad3a2e0356b65fdfe4a73c825b78071ae469db35162978518b8c258abb3767: 1
0x00058a56ea94653cdf4f152d227ace22d4c00ad99e2a43f58cb7d9e3feb295f2: 4
0x2b627736bca15cd5381dcf80b0bf11fd197d01a037c52b927a881a10fb73ba61: 5
0xb3d084820fb1a9decffb176436bd02558d15fac9b0ddfed8c465bc7359d7dce0: 8
0x0d96f072d487d7f8b4891a4e4cf14e8cdad444a34248230085c20808d57caa1a: 3
0x5be70b68c8361762980ec7d425d79fd33f6d49cac8a498e6ddf514f995b987f7: 1
0x984b5f16f61de82e7fa1d8fea81be585cd2b484ac6391b16741356cf4b7393d1: 1
0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925: 28
0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1: 31
0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f: 13
0xda919360433220e13b51e8c211e490d148e61a3bd53de8c097194e458b97f3e1: 7
0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496: 1
0x79f19b3655ee38b1ce526556b7731a20c8f218fbda4a3990b6cc4172fdf88722: 1
0xefefaba5e921573100900a3ad9cf29f222d995fb3b6045797eaea7521bd8d6f0: 1
0x458f5fa412d0f69b08dd84872b0215675cc67bc1d5b6fd93300a1c3878b86196: 13
0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c: 19
0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef: 245
0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822: 17
0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65: 3
0x14fc68fa3d99c92bb4159f5ae1ddd4bbf7b874931534c08ac40467d7ece273d6: 3
0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31: 1
Time: 0h:00m:34s     
```

## Version 2
I made the processBlockRange function to process each block in parrallel, with number of worker = 10, as an result, it now took only 6 seconds to process;
```
Total event counts across the block range:
0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f: 13
0x79f19b3655ee38b1ce526556b7731a20c8f218fbda4a3990b6cc4172fdf88722: 1
0x5be70b68c8361762980ec7d425d79fd33f6d49cac8a498e6ddf514f995b987f7: 1
0x0c5bc74ccdf848b38eb526a154b85085e1d61addf1d100cba2074e039c0b6340: 4
0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925: 28
0x14fc68fa3d99c92bb4159f5ae1ddd4bbf7b874931534c08ac40467d7ece273d6: 3
0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822: 17
0x00058a56ea94653cdf4f152d227ace22d4c00ad99e2a43f58cb7d9e3feb295f2: 4
0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31: 1
0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496: 1
0xda919360433220e13b51e8c211e490d148e61a3bd53de8c097194e458b97f3e1: 7
0x02ad3a2e0356b65fdfe4a73c825b78071ae469db35162978518b8c258abb3767: 1
0xb3d084820fb1a9decffb176436bd02558d15fac9b0ddfed8c465bc7359d7dce0: 8
0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef: 245
0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c: 19
0x2b627736bca15cd5381dcf80b0bf11fd197d01a037c52b927a881a10fb73ba61: 5
0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65: 3
0x984b5f16f61de82e7fa1d8fea81be585cd2b484ac6391b16741356cf4b7393d1: 1
0xefefaba5e921573100900a3ad9cf29f222d995fb3b6045797eaea7521bd8d6f0: 1
0x0d96f072d487d7f8b4891a4e4cf14e8cdad444a34248230085c20808d57caa1a: 3
0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1: 31
0x458f5fa412d0f69b08dd84872b0215675cc67bc1d5b6fd93300a1c3878b86196: 13
0x804c9b842b2748a22bb64b345453a3de7ca54a6ca45ce00d415894979e22897a: 13
Time: 0h:00m:06s   
```

## Setup
Refer to Makefile

### Database related
```
make postgres
make createdb
make migrate
```
### Server
```
make docker_build
```

OR 
```
make server
```

Then run the tests:
```
cd test
go test -v -run TestAPISuite
```
