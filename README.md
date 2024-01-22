# Cooker 👨🏻‍🍳

Cooker is a testing framework designed to automatically generate test suites for API endpoints. The framework utilizes Go (Golang) to streamline the process of creating and running tests for RESTful APIs. 
Below is an overview of the framework, including installation instructions, an outline of the main services, and an explanation of the Cooker folder structure with rules for generating endpoints.

# Installation 🔥

To use Cooker, follow these installation steps:
```bash
git clone https://github.com/mariobenissimo/Cooker.git 
```

or add it 
```bash
go get github.com/mariobenissimo/Cooker.git 
```
# Usage 🔪

To use Cooker inside the project use:
```bash
Cooker.CreateCooker(port: ":8082")
```
to use dashboard needed to write json

or use
```bash
Cooker.cook(path: "../cooker")
```
placing the json files in a folder is to pass it as a parameter

# Framework Overview 🍽️
Cooker is designed to streamline the process of testing API endpoints. It leverages Docker for containerization, GitHub Actions for continuous integration, and Go for writing and executing tests. The main services include:

* Cooker Service: Responsible for generating test cases based on JSON descriptions of API endpoints.
* Server Service: A dummy server that manages a user database and exposes various API routes.
* API Gateway Service: Provides a single entry point for all APIs and uses Gorilla Mux for routing.

# Rules for Generating Endpoints 🎛️
Cooker generates tests based on JSON descriptions and follows specific rules:

* Authentication Rule: If the description includes the "authentication" key, the test generates JWT-based authentication using a secret token.

* Rate Limiter Rule: If the server uses a rate limiter, the test includes scenarios to check its functionality. Parameters include maxRequests and seconds to control the rate of requests.

* HTTP Status Rule:Cooker ensures that the generated tests cover various HTTP status codes, such as 200, 201, 400, 401, and 429.

# Example 🍝

Example of JSON description are avaibles under the folder 'cooker', testing are avaibles under the folder 'testing'

# Contribution 💪🏻

Feel free to explore Cooker, customize the JSON descriptions, and adapt the rules to suit your specific API testing needs. If you want contribute to the project, add your logic in the comment of the pull request and create a new folder under cooker, you can use ast block presents in other folder or create new ast block to generate code. If you encounter any issues, refer to the GitHub repository for additional documentation and support. 
Create new recipes and happy testing with Cooker!
