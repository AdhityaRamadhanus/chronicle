# chronicle

[![Go Report Card](https://goreportcard.com/badge/github.com/AdhityaRamadhanus/chronicle)](https://goreportcard.com/report/github.com/AdhityaRamadhanus/chronicle)  [![Build Status](https://travis-ci.org/AdhityaRamadhanus/chronicle.svg?branch=master)](https://travis-ci.org/AdhityaRamadhanus/chronicle)

Experimental golang microservices rest api without orm on news domain

Entities:
Topics
Stories

Database: postgres
cache-server: redis

Authentication: JWS
<p>
  <a href="#installation-for-development">Installation |</a>
  <a href="#Usage">Usage |</a>
  <a href="#licenses">License</a>
  <br><br>
  <blockquote>
	chronicle is rest api microservices about news.

  This project is only for experiment. There is much work to do for this project to be complete.
  </blockquote>
</p>

Installation (For Development)
----------- 
* git clone
* set environt variables in .env (example below)
* change config if necessary in config/development/config.yml
```
ENV=development

PORT=8000

PG_DATA_DIR=/home/path/to/your/postgresql-container-data

// leave this for production
PRODUCTION_JWT_SECRET=
PRODUCTION_CACHE_RESPONSE=

PRODUCTION_LOGGLYTOKEN=
PRODUCTION_LOGGLYHOST=

PRODUCTION_REDIS.HOST=
PRODUCTION_REDIS.PORT=
PRODUCTION_REDIS.PASSWORD=
PRODUCTION_REDIS.DB=
```
* docker-compose up
* create database "chronicle" on postgres
* create database "chronicle-test" on postgres
* run migration 
``` bash
make migration
```
* run build
```bash
make
```
* run chronicle

Usage
-----
* You will need access token to use the api
* generate access token
```bash
make generate-token
```

License
----

MIT © [Adhitya Ramadhanus]

