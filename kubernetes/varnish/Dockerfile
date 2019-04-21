FROM cooptilleuls/varnish:6.2-alpine

# Install vmods

ENV VMOD_QUERYSTRING_VERSION 2.0.1
ENV VMOD_QUERYSTRING_SHA256 34540b0fb515bfbf9aaa4154be5372ce5aa8c7050f35f07dc186c85bb7e976c0

ENV VMOD_DYNAMIC_VERSION b8731c42f73075a112d4b3475c1da08a5e85fcec
ENV VMOD_DYNAMIC_SHA256 c70fb00e5763d8cd460c4fa6c7ae68cf74d0ba89ff4393f8e5b34d316cd18aa1

RUN set -eux; \
	\
	fetchDeps=' \
		ca-certificates \
		wget \
	'; \
	buildDeps=" \
		$VMOD_BUILD_DEPS \
    dpkg \
		dpkg-dev \
    py-docutils \
    pcre-dev \
    libexecinfo-dev \
	"; \
	apk add --no-cache --virtual .build-deps $fetchDeps $buildDeps; \
	\
	wget -O vmod-querystring.tar.gz "https://github.com/Dridi/libvmod-querystring/releases/download/v$VMOD_QUERYSTRING_VERSION/vmod-querystring-$VMOD_QUERYSTRING_VERSION.tar.gz"; \
  echo "$VMOD_QUERYSTRING_SHA256 *vmod-querystring.tar.gz" | sha256sum -c -; \
  wget -O vmod-dynamic.tar.gz "https://github.com/nigoroll/libvmod-dynamic/archive/$VMOD_DYNAMIC_VERSION.tar.gz"; \
  echo "$VMOD_DYNAMIC_SHA256 *vmod-dynamic.tar.gz" | sha256sum -c -; \
  \
  mkdir -p /usr/local/src/vmod-querystring; \
  mkdir -p /usr/local/src/vmod-dynamic; \
	tar -zxf vmod-querystring.tar.gz -C /usr/local/src/vmod-querystring --strip-components=1; \
  tar -zxf vmod-dynamic.tar.gz -C /usr/local/src/vmod-dynamic --strip-components=1; \
	rm vmod-querystring.tar.gz; \
  rm vmod-dynamic.tar.gz; \
	\
  gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)"; \
	cd /usr/local/src/vmod-querystring; \
	./configure \
		--build="$gnuArch" \
	; \
	make -j "$(nproc)"; \
	make install; \
  \
  cd /usr/local/src/vmod-dynamic; \
  ./autogen.sh; \
	./configure \
		--build="$gnuArch" \
	; \
	make -j "$(nproc)"; \
	make install; \
  \
	cd /; \
	rm -rf /usr/local/src/vmod-querystring; \
  rm -rf /usr/local/src/vmod-dynamic; \
	\
	apk del .build-deps
