package container

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8SPodController(imageRegistry *ImageRegistry, namespace string) (*K8SPodController, error) {
	c, err := K8SClientsetInCluster()
	if err != nil {
		c, err = K8SClientsetInLocal()
		if err != nil {
			return nil, err
		}
	}

	return &K8SPodController{
		client:        c,
		imageRegistry: imageRegistry,
		namespace:     namespace,
	}, nil
}

type K8SPodController struct {
	namespace     string
	client        *kubernetes.Clientset
	imageRegistry *ImageRegistry
}

func (c *K8SPodController) Apply(ctx context.Context, name string, container *Container) error {
	if err := c.ensureImagePullSecret(c.namespace); err != nil {
		return err
	}

	if err := c.applyDeployment(c.namespace, convertContainerToDeployment(name, container, c.imageRegistry)); err != nil {
		return err
	}

	return nil
}

func (c *K8SPodController) Kill(ctx context.Context, name string) error {
	return c.deleteDeployment(c.namespace, name)
}

func convertContainerToDeployment(name string, c *Container, imageRegistry *ImageRegistry) *appsv1.Deployment {
	d := &appsv1.Deployment{}
	d.Name = name

	d.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"pipeline/stage": d.Name,
		},
	}

	d.Annotations = c.Annotations

	d.Spec.Replicas = &c.Replicas

	d.Spec.Template.Labels = map[string]string{
		"pipeline/stage": d.Name,
	}

	d.Spec.Replicas = &c.Replicas

	podContainer := corev1.Container{}

	podContainer.Name = d.Name
	podContainer.Image = imageRegistry.Fix(c.Image)
	podContainer.Command = c.Command
	podContainer.Args = c.Args
	podContainer.ImagePullPolicy = corev1.PullAlways

	for k, v := range c.Envs {
		podContainer.Env = append(podContainer.Env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	d.Spec.Template.Spec.Containers = []corev1.Container{podContainer}
	d.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{{
		Name: imageRegistry.Name,
	}}

	return d
}

func (c *K8SPodController) applyDeployment(namespace string, deployment *appsv1.Deployment) error {
	api := c.client.AppsV1().Deployments(namespace)

	_, err := api.Get(deployment.Name, metav1.GetOptions{})
	if err != nil {
		if isKubeNotFound(err) {
			_, err := api.Create(deployment)
			return err
		}
		return err
	}

	_, errForUpdate := api.Patch(deployment.Name, types.MergePatchType, MustJSONContent(deployment))
	if errForUpdate != nil {
		return errForUpdate
	}
	return nil
}

func (c *K8SPodController) ensureImagePullSecret(namespace string) error {
	api := c.client.CoreV1().Secrets(namespace)

	if _, err := api.Get(c.imageRegistry.Name, metav1.GetOptions{}); err != nil {
		if isKubeNotFound(err) {
			_, err := api.Create(secretFromImageRegistry(c.imageRegistry))
			return err
		}
		return err
	}

	return nil
}

func (c *K8SPodController) deleteDeployment(namespace string, name string) error {
	api := c.client.AppsV1().Deployments(namespace)

	if err := api.Delete(name, &metav1.DeleteOptions{}); err != nil {
		if isKubeNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func secretFromImageRegistry(imageRegistry *ImageRegistry) *v1.Secret {
	secret := &v1.Secret{}
	secret.Type = v1.SecretTypeDockerConfigJson
	secret.Name = imageRegistry.Name
	secret.Data = map[string][]byte{
		".dockerconfigjson": imageRegistry.DockerConfigJSON(),
	}
	return secret
}

func isKubeNotFound(err error) bool {
	statusErr, ok := err.(*k8serrors.StatusError)
	return ok && statusErr.ErrStatus.Code == http.StatusNotFound
}

// see more https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#accessing-the-api-from-a-pod
func K8SClientsetInCluster() (*kubernetes.Clientset, error) {
	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientConfig.Wrap(newWrapTransportForLog())
	return kubernetes.NewForConfig(clientConfig)
}

func K8SClientsetInLocal() (*kubernetes.Clientset, error) {
	clientConfig, err := K8SLocalConfig()
	if err != nil {
		return nil, err
	}
	clientConfig.Wrap(newWrapTransportForLog())
	return kubernetes.NewForConfig(clientConfig)
}

func MustJSONContent(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func K8SLocalConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	apiConfig, err := rules.Load()
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
}

func newWrapTransportForLog() func(rt http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &logRoundTripper{rt: rt}
	}
}

type logRoundTripper struct {
	rt http.RoundTripper
}

func (logRoundTripper *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return logRoundTripper.rt.RoundTrip(req)
	}

	startedAt := time.Now()

	response, err := logRoundTripper.rt.RoundTrip(req)

	defer func() {
		logrus.WithContext(req.Context()).WithFields(logrus.Fields{
			"cost":   time.Since(startedAt).String(),
			"method": req.Method,
			"path":   req.URL.String(),
		}).Info()
	}()

	return response, err
}
