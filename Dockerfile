FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-alpine
COPY tracking-api /tracking-api
ENTRYPOINT ["/tracking-api"]
