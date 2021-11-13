# mockarena
A testing utility to mock out multiple network services and report on their usage

## Example
```yaml
port: 8080
mocks:
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
