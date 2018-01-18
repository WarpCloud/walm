import flask_marshmallow
from apispec.ext.marshmallow.swagger import fields2jsonschema, field2property
from flask_restplus.model import Model as OriginalModel
from werkzeug import cached_property


class Model(OriginalModel):
    def __init__(self, name, model, **kwargs):
        # XXX: Wrapping with __schema__ is not a very elegant solution.
        super(Model, self).__init__(name, {'__schema__': model}, **kwargs)

    @cached_property
    def __schema__(self):
        schema = self['__schema__']
        if isinstance(schema, flask_marshmallow.Schema):
            return fields2jsonschema(schema.fields)
        elif isinstance(schema, flask_marshmallow.base_fields.FieldABC):
            return field2property(schema)
        raise NotImplementedError()
