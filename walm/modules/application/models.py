# encoding: utf-8

from walm.exception import ObjectDoesNotExist
from walm.extensions import db


class Application(db.Model):

    __tablename__ = 'application'

    id = db.Column(db.String(128), primary_key=True, nullable=False)
    name = db.Column(db.String(128), nullable=False)

    @classmethod
    def get(cls, application_id=None, **kwargs):
        """Get specific application or a general query"""
        if application_id is not None:
            application = cls.query.filter_by(id=application_id).first()
            if application is None:
                raise ObjectDoesNotExist('Application "%s" does not exists' % application_id)
            return application

        return cls.query.filter_by(**kwargs).all()
