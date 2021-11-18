# mockarena
A testing utility to mock out multiple network services and report on their usage

## Example
```yaml
port: 8080
mocks:
  - name: "db"
     type: "mysql"
     port: 4000
     databases:
      - name: "main_db"
        returnSequence:
          - result:
              rowsAffected: 2
              insertID: 3
            repeat:
              for: 1m
          - error:
              sqlError:
                num: 0
                state: ""
                message: ""
          - rows:
              fields:
                - name: "field_name"
                  type: "DATE"
                  table: "students"
                  charset: "latin1"
                  decimals: "?"
                  flags: ""
              sessionStateChanges: "?"
              statusFlagsRaw: 3
              rows:
                -
                  - "2021-10-20"
                -
                  - "2021-10-21"
            repeat:
              forever: nonblocking
   - name: "web_server"
     type: "HTTP"
     port: 4000
     serial: true
     record: ["address", "header", "time", "body"]
     paths:
      - path: "/"
        methods:
          - method: "GET"
            returnSequence:
              - &default_http
                header: &default_header
                  - key: "Content-Type"
                    value: "text/plain"
                body: "OK"
                delay: 500ms
                status: 200
                repeat:
                  # until: "2021-11-01 16:55:00"
                  # for: 30s
                  count: 2
              - delay: 3s
                proxy:
                  host: "http://kraft.singles"
          - method: "POST"
            returnSequence:
              - header: *default_header
                body: "ERROR"
                status: 500
                repeat:
                  forever: "nonblocking"
      - path: "/testpath"
        methods:
        - returnSequence:
          - <<: *default_http
            body: "TEST PATH"
            repeat:
              for: 10s
```
