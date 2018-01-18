from flask_sqlalchemy import SQLAlchemy

db = SQLAlchemy(session_options={'autocommit': True})


def init_app(app, **kwargs):
    """
    Application extensions initialization.
    """
    db.init_app(app)
