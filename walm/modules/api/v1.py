from flask_restplus import Api

api = Api(
    version='1.0',
    title="WALM API",
    description=(
        "Warp Application Lifecycle Management service\n"
    ),
)

api_prefix = '/api/v1'
