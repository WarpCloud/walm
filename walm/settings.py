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
import os


class Settings(object):
    DEBUG = False
    SQLALCHEMY_TRACK_MODIFICATIONS = False

    # Flask restplus settings
    RESTPLUS_SWAGGER_UI_DOC_EXPANSION = 'list'
    RESTPLUS_VALIDATE = True
    RESTPLUS_MASK_SWAGGER = False
    RESTPLUS_ERROR_404_HELP = False

    PROJECT_ROOT = os.path.abspath(os.path.dirname(__file__))

    DEFAULT_DATABASE_URI = 'sqlite:///%s' % (os.path.join(PROJECT_ROOT, "walm.db"))
    SQLALCHEMY_DATABASE_URI = os.getenv('SQLALCHEMY_DATABASE_URI', DEFAULT_DATABASE_URI)

    ENABLED_MODULES = (
        'application',
        'api'  # Keep api module as last one for module injection
    )

    # WALM web server settings
    WALM_WEBSERVER_ADDRESS = os.getenv('WALM_WEBSERVER_ADDRESS', '0.0.0.0')
    WALM_WEBSERVER_PORT = os.getenv('WALM_WEBSERVER_PORT', '6180')
    WALM_WORKERS = os.getenv('WALM_WORKERS', 5)
    WALM_WEBSERVER_TIMEOUT = os.getenv('WALM_WEBSERVER_TIMEOUT', '3600')

    PRODUCT_META_HOME = os.getenv('PRODUCT_META_HOME', 'product-meta')


class ProductionConfig(Settings):
    DEBUG = False


class DevelopmentConfig(Settings):
    DEBUG = True


class TestingConfig(Settings):
    TESTING = True
    # Use in-memory SQLite database for testing
    SQLALCHEMY_DATABASE_URI = 'sqlite://'
