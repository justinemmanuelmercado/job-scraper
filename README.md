# Job Scraper for workfindy.com

## Description
This is a job scraper for workfindy.com. It scrapes the job title, company name, location, and job description. It then saves the data to a DB


## Installation
0. Install Go and Postgres
1. Clone the repo
   ```sh
   git clone
    ```
2. Fill in the .env file with your credentials (Reference the .env.example file)


## Usage
1. Run the main.go file
   ```sh
   go run main.go
   ```
2. The data should be saved to the DB -- Because I can't be arsed for safety, errors just get logged and the program continues to run.
3. A sample prisma.schema file is included for reference on how to create the DB

## License
Distributed under the MIT License. See `LICENSE` for more information.
