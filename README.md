## Tunnel for secure connection between services

### Usage
```
// Connect to MySQL database
import "github.com/PumpkinSeed/tunnel"
.
.
.
rConn := tunnel.New("localhost", "192.168.42.10", "localhost", 3307, 2203, 3306)
rConn.AuthWithRSAKey("root", "/home/user/.ssh/id_rsa")
go rConn.Setup()
```
After you can connect to the remote database like `localhost:3307`