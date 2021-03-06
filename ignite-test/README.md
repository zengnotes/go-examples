## Ignite Test

A simple program to throw some load at ignite to simulate anticipated access patterns.

### Amazon Setup - Cluster Nodes

Note - for a long running test on ec2, I saw the cluster members run out
of memory on micro instances -- t2.small instances worked fine.

Install docker and pull the grid gain image.

<pre>
wget -qO- https://get.docker.com/ | sh
sudo usermod -aG docker ubuntu
docker pull gridgain/gridgain-com
</pre>

Note you need to logout then log back in to pick up the docker group.

Next pull in the config and edit the particulars.

<pre>
wget https://raw.githubusercontent.com/d-smith/go-examples/master/ignite-test/aws-cache-config.xml
</pre>

Note in the above, the security groups need to permit traffic on ports 47100 and 47500 amongst the
cluster members.

To run the cache:

<pre>
docker run -it --net=host -v /home/ubuntu:/ignite -e "CONFIG_URI=file:///ignite/aws-cache-config.xml" -e "OPTION_LIBS=ignite-rest-http"  -p 8080:8080 -e "IGNITE_QUIET=false" gridgain/gridgain-com
</pre>

### Amazon Setup - Test Nodes

For nodes that will run the test, set up docker, git, and golang

<pre>
wget -qO- https://get.docker.com/ | sh
sudo usermod -aG docker ubuntu
docker pull golang:latest
</pre>

Now pull down the client and compile it - note we start the golang container as a terminal session and
run the build commands in a bash shell in the container.

<pre>
wget https://raw.githubusercontent.com/d-smith/go-examples/master/ignite-test/ignition.go
docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.6 bash
 go get github.com/alecthomas/kingpin
 go get github.com/nu7hatch/gouuid
 go build -v -o ignitor
 exit
</pre>

Finally run the test client - use the `--help` option to get the command line options.

Note for long running tests you should start all of the above using screen and attach/detach/reattach as appropriate.
Alternatively for the cache clusters you could run the docker processes as daemons instead of interactively as shown
in this README.


