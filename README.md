# delay-tasks

An implementation of delayed tasks.

![diag-1](delay-tasks.jpg)

## Usage

```bash
$ git clone https://github.com/rosbit/delay-tasks
$ cd delay-tasks
$ make
```

  An executable `delay-tasks` will be generated. Run it like the following:

```bash
$ CONF_FILE=./sample.conf.json ./delay-tasks
```

## API

### 1. Register category handler

- POST /handler/:cate

- Body:
  
  ```json
  {"handler": "a http url"}
  ```

- Response
  
  ```json
  {
     "code": 200,
     "msg": "OK"
  }
  ```

- A handler must be implemented as an endpoint satisfy the following:
  
  - method: POST
  - BODY: a JSON, see params detail in Create/Update task.
    ```json
    {
       "cate": "task category",
       "key": uint64,  // task key in the category
       "params": JSON, // params when adding task
       "inAdvance": false|true // whether the task is run in advance.
    }
    ```

### 2. Create/Update a delayed task in a category

- POST /task/:cate

- Body:
  
  ```json
  {
    "timestamp": in-seconds,
    "key": uint64,
    "params": {anything},
    "handler": "using caterory handler if it is blank"
  }
  ```

- Response
  
  ```json
  {
     "code": 200,
     "msg": "OK"
  }
  ```

### 3. Remove a delayed task in a category

- DELETE /task/:cate

- Body:
  
  ```json
  {
     "timestamp": in-seconds,
     "key": uint64, 
     "exec": true|false // executing task before being removed if true.
  }
  ```

- Response
  
  ```json
  {
     "code": 200,
     "msg": "OK"
  }
  ```

### 4. Get task

- GET /task/:cate/:key[?timestamp=xxx]

- Response
  
  ```json
  {
     "code": 200,
     "msg": "OK",
     "result": {
        "timeToRun": xxxx,
        "params": {JSON},
        "handler": "http://handler"
     }
  }
  ```

### 5. List tasks

- GET /tasks
- Result will be dumped as Response.
