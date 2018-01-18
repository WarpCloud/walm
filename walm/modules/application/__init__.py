from walm.modules.api import api_current


def init_app(app, **kwargs):
    from . import models, resources
    api_current.api.add_namespace(resources.api)
