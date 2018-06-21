local app = import '../../../applib/app.libsonnet';
local kube = import '../../../applib/kube.libsonnet';
{
    walm:: {
        deployment(_name, config, env)::kube.tos.Deployment(
            kube['extensions/v1beta1'].Deployment(name=_name + '-', moduleName=_name, config=config) {
                spec+: {
                    replicas: config.App.replicas,
                    template+: {
                        metadata+: {
                            annotations +: kube.NodeAntiAffinityAnnotations(config, moduleName=_name),},
                            spec+: {
                                terminationGracePeriodSeconds: 0,
                                priority: config.App.priority,
                                containers: [
                                    kube.v1.PodContainer(_name) {
                                        image: config.App.image,
                                        imagePullPolicy: 'Always',
                                        args: ['walm serv'],
                                        env: env ,
                                        resources: kube.v1.ContainerResourcesV2(config.App.resources),
                                        },
                                        ],
                                initContainers: [
                                    {
                                        name: 'walm-init',
                                        image: config.App.image,
                                        imagePullPolicy: 'Always',
                                        env: env,
                                        command: ['walm migrate'],
                                    },
                                ],
              },
            },
          },
        },
        config
      ),

    svc(_name, config)::kube.v1.NodePortService(name=_name + '-', moduleName=_name, config=config) {
        spec+: {
          ports: [
            { name: 'port', port: 8000, protocol: 'TCP', targetPort: 8000 },
          ],
        },
      },

    headless_svc(_name, config)::
      kube.v1.HeadlessService(name=_name + '-hl-', moduleName=_name + '-hl', selectorModuleName=_name, config=config) {
        metadata+: {
          annotations+: {
            'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
          },
        },
        spec+: {
          ports: [
            { name: 'walm-port', port: 8000, protocol: 'TCP' },
          ],
        },
      },
    },

}
