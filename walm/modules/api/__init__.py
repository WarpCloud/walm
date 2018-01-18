from flask import Blueprint

from walm.modules.api import v1

api_current = v1


def init_app(app, **kwargs):
    blueprint = Blueprint('api', __name__, url_prefix=api_current.api_prefix)
    api_current.api.init_app(blueprint)
    app.register_blueprint(blueprint)
