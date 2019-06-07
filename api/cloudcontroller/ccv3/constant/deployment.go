package constant

// DeploymentState describes the states a zero down time deployment used to
// push new apps without restart can be in.
type DeploymentState string

const (
	// DeploymentDeploying means the deployment is in state 'DEPLOYING'
	DeploymentDeploying DeploymentState = "DEPLOYING"

	// DeploymentCanceled means the deployment is in state 'CANCELED'
	DeploymentCanceled DeploymentState = "CANCELED"

	// DeploymentDeployed means the deployment is in state 'DEPLOYED'
	DeploymentDeployed DeploymentState = "DEPLOYED"

	// DeploymentCanceled means the deployment is in state 'CANCELING'
	DeploymentCanceling DeploymentState = "CANCELING"

	// DeploymentFailing means the deployment is in state 'FAILING'
	DeploymentFailing DeploymentState = "FAILING"

	// DeploymentFailed means the deployment is in state 'FAILED'
	DeploymentFailed DeploymentState = "FAILED"
)
