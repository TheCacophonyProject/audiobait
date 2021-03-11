#!/bin/bash
systemctl daemon-reload
systemctl enable audiobait.service
systemctl restart audiobait.service