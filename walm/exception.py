class WALMException(Exception):
    """
    Base exception for all Kubernetes errors.
    """

    def __init__(self, message, error_code=None):
        super(WALMException, self).__init__(message)
        self.message = message
        self.error_code = error_code


class ObjectDoesNotExist(WALMException):
    def __init__(self, message, status_code=None):
        super(ObjectDoesNotExist, self).__init__(message, status_code)
