from walm.extensions.flask_restplus.schema import ModelSchema
from .models import Application


class ApplicationSchema(ModelSchema):
    class Meta(ModelSchema.Meta):
        model = Application
        dump_only = (
            Application.id.key,
            Application.name.key
        )
