FROM 172.16.1.99/transwarp/walm_builder:tdc-1.0

MAINTAINER xiaming.chen@transwarp.io

ENV WALM_HOME=/usr/lib/walm
ENV WALM_LOG_DIR=/var/log/walm
ENV WALM_TMP=/tmp/walm
ENV WALM_WEBSERVER_PORT=6180

WORKDIR $WALM_TMP
ADD . $WALM_TMP

RUN pip install -U -i http://172.16.1.161:30033/repository/pypi/simple/ \
    -r $WALM_TMP/requirements.txt \
    --trusted-host=172.16.1.161 \
    --timeout=120

RUN mkdir -p $WALM_LOG_DIR
RUN chmod -R 777 $WALM_LOG_DIR

RUN cd $WALM_TMP
RUN python setup.py install

# Install necessary files
COPY ./docker-build/scripts/ $WALM_HOME/scripts/
COPY ./migrations/ $WALM_HOME/migrations/
COPY ./bin/walm_server $WALM_HOME/scripts/

# Clear temporary files
RUN rm -rf $WALM_TMP

# Make binaries executable and add to global PATH
RUN chmod +x $WALM_HOME/scripts/*
ENV PATH $PATH:$WALM_HOME/scripts/

EXPOSE $WALM_WEBSERVER_PORT
