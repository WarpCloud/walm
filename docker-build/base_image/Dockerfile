FROM 172.16.1.99/transwarp/tdh-baseimage:transwarp-5.2

MAINTAINER xiaming.chen@transwarp.io

# Set transwarp repo
RUN yum clean all
RUN rm /etc/yum.repos.d/* && \
    echo -e "[transwarp]\nname = Linux OS Packages\nbaseurl = ftp://172.16.1.32/pub/centos7-1511-everything\ngpgcheck = 0" \
        > /etc/yum.repos.d/transwarp.repo

RUN yum install -y wget make gcc gcc-c++ git mysql-devel python-devel python-setuptools python-pip python-wheel unixODBC-devel

RUN yum install -y unixODBC unixODBC-devel MySQL-python python-setuptools python-devel gcc-c++ python-pip which mysql git

RUN yum clean all && rm -Rf /var/cache/yum

RUN pip install -U setuptools
RUN pip install --upgrade pip
EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
