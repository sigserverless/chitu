# Bash Runtime 

## AWS Lambda Custom Runtime 

```shell
bash-runtime $ chmod 755 function.sh bootstrap
bash-runtime $ zip function.zip function.sh bootstrap

bash-runtime $ aws lambda create-function --function-name bash-runtime \
--zip-file fileb://function.zip --handler function.handler --runtime provided \
--role arn:aws:iam::xxxxxxxxxxxx:role/lambda-role
```

## Layer

```shell
$ zip runtime.zip bootstrap

$ aws lambda publish-layer-version --layer-name bash-runtime --zip-file fileb://runtime.zip

$ aws lambda update-function-configuration --function-name bash-runtime \
--layers arn:aws:lambda:us-east-1:xxxxxxxxxxxx:layer:bash-runtime:1
```

## Update Function

```shell
$ zip function-only.zip function.sh

$ aws lambda update-function-code --function-name bash-runtime --zip-file fileb://function-only.zip
```

## Invoke 

```shell
$ aws lambda invoke --function-name bash-runtime --payload '{"text":"Hello"}' response.txt --cli-binary-format raw-in-base64-out

$ cat response.txt
```

# Coordinator

## Deploy

```shell
zip function.zip .
```