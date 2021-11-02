# mockarena
A testing utility to mock out multiple network services and report on their usage

## Example
```yaml
reportPath: "./out.json"
port: 8080
mocks:
   - name: "web_server"
     type: "HTTP"
     port: 4000
     serial: true
     record: ["address", "header", "time", "body"]
     returnSequence:
        - &default_http
          header: &default_header
            - key: "Content-Type"
              value: "text/plain"
          body: "OK"
          delay: 500ms
          status: 200
          repeat:
            until: "2021-11-01 16:55:00"
            for: 30s
            count: 2
        - header: *default_header
          body: "ERROR"
          status: 500
        - <<: *default_http
          repeat:
            for: 10s
```
