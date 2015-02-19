FROM daaku/goruntime
COPY ouiserver-linux-amd64 /ouiserver

EXPOSE 5000

CMD ["/ouiserver", "-open=http", "-update-every=@daily"]
