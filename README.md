Simple http server that can be used as a multitool for testing and learning.
* "/req"      - sends request to external service and retuns its response; use ?url=
* "/error"    - returns error 500
* "/error2"   - returns error 500 every second request
* "/host"     - returns hostname
* "/ip"       - returns ip address of the host
* "/env"      - returns env variables
* "/headers"  - returns headers
* "/hello"    - returns hostname, ip address and vslue of RETURN_TEXT env * variable if available
* "/ls"       - returns directory contents; use ?path=PATH to select a directory
* "/"         - returns hostname, ip address and value of RETURN_TEXT env variable if available

```
kubectl run multitool --image=przemekmalak/multitool --env="RETURN_TEXT=service 1" --port 8080

```