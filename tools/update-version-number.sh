#!/bin/bash

# 檢查參數數量
if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo "Usage: $0 <new_version> [--update_ui]"
    exit 1
fi

if [ ! -z "$2" ]; then
    UPDATE_UI="true"
else
    UPDATE_UI="false"
fi

CHART_YAML="./.ci/helm/Chart.yaml"
CHART_DEFAULT_VALUES="./.ci/helm/values.yaml"
PACKAGE_JSON="./package.json"
UI_PACKAGE_JSON="./ui/package.json"
NEW_VERSION="$1"

# 使用 sed 來替換 Chart.yaml的 version 行
sed -i -e "s/^\(version:\s*\).*/\1$NEW_VERSION/" $CHART_YAML
echo "Updated version in $CHART_YAML to $NEW_VERSION."

# 使用 sed 來替換 values.yaml 的 image 行
sed -i -e "s/^\(\s*image:\s*.*:\).*/\1$NEW_VERSION/" $CHART_DEFAULT_VALUES
echo "Updated version in $CHART_DEFAULT_VALUES to $NEW_VERSION."

# 使用 sed 來替換 package.json 的 version 行
sed -i -e "s/^\(\s*\"version\":\s*\)[^,]*/\1\"$NEW_VERSION\"/" $PACKAGE_JSON
echo "Updated version in $PACKAGE_JSON to $NEW_VERSION."

# 如果設定要更新UI版本，則更新ui/package.json
if [ "$UPDATE_UI" = "true" ]; then
    sed -i -e "s/^\(\s*\"version\":\s*\)[^,]*/\1\"$NEW_VERSION\"/" $UI_PACKAGE_JSON
    npm -w ui install
    echo "Updated version in $UI_PACKAGE_JSON to $NEW_VERSION."
fi