#!/usr/bin/env python3
# -*- coding: utf-8 -*-
""" standard imports """
import yaml
import json
import logging
import sys
import os
import re
import requests
from typing import Tuple
from app.api import Api
from datetime import datetime

# Defense Unicorns
# start and finish is same, date ran.

# Create a new assessment as a child of the control on the ISTIO component

# Show the assessment as passing with the plan and results

# Use automation info to store some data on the assessment

# self update GRC in real time.

logger = logging
logger.basicConfig(level=logging.INFO)


def dump_yaml(data: dict, file_name: str):
    """Create Yaml File"""
    # assert plan_name, user_id, config['confidentiality'], config['integrity'], config['availability'], config['systemType'], config['overallCategorization'], config['isPublic']
    if not os.path.exists("init.yaml"):
        try:
            with open(file_name, "w") as f:
                yaml.dump(data, f)
                logger.info(
                    "Init.yaml file created, please edit with desired values and re-run this application."
                )
                sys.exit()
        except yaml.YAMLError as ex:
            logger.error("YAMLError! \n%s", ex)
            sys.exit()


def get_conf() -> dict:
    """Get configuration from init.yaml if exists"""
    fname = "init.yaml"
    data = {"host": "https://dev.regscale.com", "user": "sam", "password": "hunter2"}
    # load the config from YAML
    try:
        with open(fname, encoding="utf-8") as stream:
            config = yaml.safe_load(stream)
    except FileNotFoundError as ex:
        logger.error(
            "FileNotFoundError: Never fear, we will create the init.yaml \n%s", ex
        )
    finally:
        dump_yaml(data, fname)
    if all(item in list(config.keys()) for item in list(data.keys())):
        # Make sure the keys are there, or complain to the user
        return config
    else:
        raise ValueError(
            f"Please make sure the following keys are in the init.yaml file:\t{list(data.keys())}"
        )


def regscale_login(url_login: str) -> Tuple[str, str]:
    """Login to RegScale"""

    # login and get token
    user_id: str = ""
    jwt: str = ""
    response = None
    try:
        response = api.post(url=url_login, headers={}, json=auth)
        auth_response = response.json()
        user_id = auth_response["id"]
        jwt = "Bearer " + auth_response["auth_token"]
        logging.debug(user_id)
        logging.debug(jwt)
        if response.raise_for_status():
            logger.error("Unable to login..")
            sys.exit()
        return user_id, jwt
    except requests.ConnectionError:
        logging.error("ConnectionError: Unable to log in to RegScale")
        sys.exit()
    except requests.RequestException:
        logging.error("RequestException: Unable to log in to RegScale")
        sys.exit()


def validate_url(url: str):
    """Regular Expression to validate a URL"""
    regex = re.compile(
        r"^(?:http|ftp)s?://"  # http:// or https://
        r"(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|"  # domain...
        r"localhost|"  # localhost...
        r"\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})"  # ...or ip
        r"(?::\d+)?"  # optional port
        r"(?:/?|[/?]\S+)$",
        re.IGNORECASE,
    )
    if not re.match(regex, url) is not None:
        logger.error("Host URL is not valid, Please check host URL and try again")
        sys.exit()


if __name__ == "__main__":

    filename = sys.argv[1]

    config = get_conf()
    host = config["host"]
    logger.info(f"Logging in to {host}\n ")
    validate_url(host)
    auth = {
        "userName": config["user"],
        "password": config["password"],
        "oldPassword": "",
    }
    api = Api(token="")
    user_id, jwt = regscale_login(url_login=f"{host}/api/authentication/login")
    api = Api(token=jwt)
    components = api.get(url=config["host"] + f"/api/components/getAll").json()
    assessments = api.get(url=config["host"] + f"/api/assessments/getAll").json()
    logger.info(f"fetched {len(components)} components")

    time = datetime.now()

    with open(filename, "r") as stream:
        data = yaml.safe_load(stream)
        for dat in data:
            result = dat["result"]
            reqs = dat["source-requirements"]
            control_id = dat["source-requirements"]['control-id']
            for rule in dat["source-requirements"]["rules"]:
                name = rule["name"].replace("-", " ").split("_", 1)[0]
                try:
                    component_id = [
                        comp
                        for comp in components
                        if comp["title"].lower() == name.lower()
                    ][0]["id"]
                except IndexError as iex:
                    logger.error("Failed component lookup, exiting..")
                    sys.exit(1)
                logger.info(name)
                existing_component_controls = api.get(url=config['host'] + f"/api/controlImplementation/getByParent/{component_id}/components").json()
                #logger.info(existing_component_controls)
                # Build assessment
                existing_component = [cntrl for cntrl in existing_component_controls if cntrl['controlName'] == control_id][0]
                #logger.info(existing_component)
                assessment = {
                    "leadAssessorId": user_id,
                    "title": rule["name"],
                    "assessmentType": "Script/DevOps Check",
                    "assessmentResult": result,
                    "plannedStart": time.strftime("%Y-%m-%dT%H:%M:%S"),
                    "plannedFinish": time.strftime("%Y-%m-%dT%H:%M:%S"),
                    "actualFinish": time.strftime("%Y-%m-%dT%H:%M:%S"),
                    "status": "Complete",
                    "assessmentPlan" : reqs['description'],
                    "assessmentReport": json.dumps(rule),
                    "targets": "",
                    "metadata": "",
                    "componentId": 145,
                    "controlId": existing_component['id'],
                    "parentId": existing_component['id'],
                    "parentModule": "controls",
                    "isPublic": True
                }
                #logger.info(assessment)
                r = api.post(url=config['host'] + f"/api/assessments", json=assessment)
                logger.info(f"""Successfully posted {rule["name"]} assessment.""")
                # if not [assess for assess in assessments if assess['title'] == assessment['title']]:
                #     r = api.post(url=config['host'] + f"/api/assessments", json=assessment)
                #     if not r.raise_for_status():
                #         logger.info(f"""Successfully posted {rule["name"]} assessment.""")
