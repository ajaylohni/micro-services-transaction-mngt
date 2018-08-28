#!/bin/bash

echo "Checking postgres service status..."

db_service=postgresql
mq_service=rabbit

service_status=ok

green='\033[0;32m' #Green color
red='\033[0;31m' #Red color
nc='\033[0m' #No color

echo -n "Enter postgres password : "
read -s db_pwd

test_database() {
        echo "Checking for database..."
        echo -n "Enter database name : "
        read db_name
        db_exists=`echo "$(psql -U postgres "password=${db_pwd}" -c "(SELECT EXISTS (SELECT 1 from pg_database WHERE datname='${db_name}'))")"` 
        echo $db_exists
        echo ${db_exists:19:1}
        if (( ${db_exists:19:1} == f ))
        then
                PGPASSWORD=${db_pwd} psql -U postgres -c "create database ${db_name};"
                echo -e "${green}database ${db_name} created successfully${nc}"
        else 
                echo "Database ${db_name} already existes!!!" 
        fi   
        echo "test_db.sql executing..."
        PGPASSWORD=${db_pwd} psql -d ${db_name} -U postgres -f test_db.sql
        echo -e "${green}test_db.sql executed successfully${nc}"
        echo "Tables and functions are into ${db_name} database."
}

if (( $(ps -ef | grep -v grep | grep $db_service | wc -l ) > 0 ))
then
	echo -e "\n${green}OK!${nc} PostgreSQL is Running..."
        test_database
else
	echo -e "${red}oops!${nc} PostgreSQL is not Running"
        $service_status=failed
fi

if (( $(ps -ef | grep -v grep | grep $mq_service | wc -l ) == 5 ))
then
        echo -e "${green}OK!${nc} RabbitMq is Running..."
else
        echo -e "${red}oops!${nc} RabbitMq is not Running"
        $service_status=failed
fi

if (( $service_status == ok ))
then
        echo -e "${green}Service Staus is OK!${nc}"
        echo "Starting the pg-amqp-bridge..."
        POSTGRESQL_URI="postgres://postgres:data@localhost/${db_name}" AMQP_URI="amqp://localhost//" BRIDGE_CHANNELS="pgchannel:task_queue" pg-amqp-bridge
fi