import logging

from flask_restplus import Namespace, Resource

from .schemas import ApplicationSchema

log = logging.getLogger(__name__)
api = Namespace('application', description="Application module")


@api.route('/')
class ApplicationResource(Resource):

    @api.response(ApplicationSchema(), description='Get applications')
    def get(self):
        from .models import Application
        return Application.get()
