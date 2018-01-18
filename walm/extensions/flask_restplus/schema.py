import flask_marshmallow


class SchemaMixin(object):
    def __deepcopy__(self, memo):
        # XXX: Flask-RESTplus makes unnecessary data copying, while
        # marshmallow.Schema doesn't support deepcopyng.
        return self


# Define ModelSchema when using sqlalchemy as orm
class ModelSchema(SchemaMixin, flask_marshmallow.sqla.ModelSchema):
    pass


class Schema(SchemaMixin, flask_marshmallow.Schema):
    pass


class DefaultHTTPErrorSchema(Schema):
    status = flask_marshmallow.base_fields.Integer()
    message = flask_marshmallow.base_fields.String()

    def __init__(self, http_code, **kwargs):
        super(DefaultHTTPErrorSchema, self).__init__(**kwargs)
        self.fields['status'].default = http_code
