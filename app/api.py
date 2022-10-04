#!/usr/bin/env python3
# -*- coding: utf-8 -*-
""" standard imports """

import concurrent.futures

import requests
from requests.adapters import HTTPAdapter, Retry


class Api:
    """Wrapper for interacting with the RegScale API"""

    def __init__(self, token: str):
        """_summary_

        Args:
            api_key (_type_): _description_
        """
        r_session = requests.Session()
        retries = Retry(
            total=5, backoff_factor=0.1, status_forcelist=[500, 502, 503, 504]
        )
        r_session.mount("https://", HTTPAdapter(max_retries=retries))
        r_session.verify = True  # Change to True if SSL enabled.
        self.session = r_session
        self.token = token
        self.accept = "application/json"
        self.content_type = "application/json"

    def get(self, url: str, headers: dict = None) -> requests.models.Response:
        """Get Request from RegScale endpoint.

        Returns:
            requests.models.Response: Requests reponse
        """
        if headers is None:
            headers = {
                "Authorization": self.token,
                "accept": self.accept,
                "Content-Type": self.content_type,
            }
        response = self.session.get(url=url, headers=headers)
        return response

    def delete(self, url: str, headers: dict = None) -> requests.models.Response:
        """Delete data from RegScale

        Args:
            url (str): _description_
            headers (dict): _description_

        Returns:
            requests.models.Response: _description_
        """
        if headers is None:
            headers = {
                "Authorization": self.token,
                "accept": "*/*",
            }
        return self.session.delete(url=url, headers=headers)

    def post(
        self, url: str, headers: dict = None, json: dict = None
    ) -> requests.models.Response:
        """Post data to RegScale.
        Args:
            endpoint (str): RegScale Endpoint
            headers (dict, optional): _description_. Defaults to None.
            json (dict, optional): json data to post. Defaults to {}.

        Returns:
            requests.models.Response: Requests reponse
        """
        if headers is None:
            headers = {
                "Authorization": self.token,
            }

        response = self.session.post(url=url, headers=headers, json=json)
        return response

    def put(
        self, url: str, headers: dict = None, json: dict = None
    ) -> requests.models.Response:
        """Update data for a given RegScale endpoint.
        Args:
            url (str): RegScale Endpoint
            headers (dict, optional): _description_. Defaults to None.
            json (dict, optional): json data to post. Defaults to {}.

        Returns:
            requests.models.Response: Requests reponse
        """
        if headers is None:
            headers = {
                "Authorization": self.token,
            }
        response = self.session.put(url=url, headers=headers, json=json)
        return response

    def update_server(
        self,
        url: str,
        headers: dict = None,
        json_list=None,
        method="post",
        config=None,
    ):
        """Concurrent Post or Put of multiple objects

        Args:
            url (str): _description_
            headers (dict): _description_
            dict_list (list): _description_

        Returns:
            _type_: _description_
        """
        if headers is None and config:
            headers = {"Accept": "application/json", "Authorization": self.token}
        if json_list and len(json_list) > 0:
            with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
                if method == "post":
                    result_futures = list(
                        map(
                            lambda x: executor.submit(self.post, url, headers, x),
                            json_list,
                        )
                    )
                for future in concurrent.futures.as_completed(result_futures):
                    try:
                        print("result is %s", future.result().status_code)
                    except Exception as ex:
                        print("e is %s, type: %s", ex, type(ex))
