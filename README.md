# highlander

Golang library to determine "master" in clustered scenarios for AWS Autoscaling Groups. It is not intended
as a replacement for a more sophisticated algorithms, but can be used to bootstrap the list of cluster
members (for example, for etcd), or it could be used in situations where a single instance from the 
autoscaling group is responsible to perform a certain action.