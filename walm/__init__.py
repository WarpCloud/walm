# Copyright (c) 2018 `Transwarp, Inc. <http://www.transwarp.io>`_.
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
#     * Redistributions of source code must retain the above copyright
#       notice, this list of conditions and the following disclaimer.
#     * Redistributions in binary form must reproduce the above copyright
#       notice, this list of conditions and the following disclaimer in the
#       documentation and/or other materials provided with the distribution.
#     * Neither the name of the Transwarp, Inc nor the names of its contributors
#       may be used to endorse or promote products derived from this software
#       without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
# TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
# PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
# LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
# NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
# SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
# encoding: utf-8
import logging
import logging.config as logging_config

import os
import yaml
from flask import Flask

# Logging utility
logging_conf = os.path.join(os.path.dirname(__file__), 'logging.yml')
if os.path.exists(logging_conf):
    logging_config.dictConfig(yaml.load(open(logging_conf, 'r')))
else:
    print('logging configuration file does not exist')

logger = logging.getLogger(__name__)


def create_app(flask_config_name=None, **kwargs):
    """
    Entry point to the RESTful Server application.
    """
    app = Flask(__name__, **kwargs)

    env_config_name = os.getenv('WALM_ENV_CONFIG')

    if not env_config_name and flask_config_name is None:
        flask_config_name = 'development'
    elif flask_config_name is None:
        flask_config_name = env_config_name
    else:
        if env_config_name:
            assert env_config_name == flask_config_name, (
                "WALM_ENV_CONFIG environment variable (\"%s\") and flask_config_name argument "
                "(\"%s\") are both set and are not the same." % (
                    env_config_name,
                    flask_config_name
                )
            )

    from walm.settings import ProductionConfig
    from walm.settings import DevelopmentConfig
    from walm.settings import TestingConfig

    config_name_mapper = {
        'production': ProductionConfig,
        'development': DevelopmentConfig,
        'testing': TestingConfig
    }

    try:
        app.config.from_object(config_name_mapper[flask_config_name])
    except ImportError:
        raise

    if app.debug:
        app.logger.setLevel(logging.DEBUG)

    from . import extensions
    extensions.init_app(app, **kwargs)

    from . import modules
    modules.init_app(app)

    from flask_migrate import Migrate
    from walm.extensions import db
    Migrate(app, db)

    return app


app = create_app()
