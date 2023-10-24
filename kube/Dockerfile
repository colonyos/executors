FROM docker.io/tensorflow/tensorflow:2.13.0-gpu

# Install dependencies
RUN apt-get update && DEBIAN_FRONTEND="noninteractive" TZ="Europe/Stockholm" apt-get install -y python3 python3-pip postgresql libpq-dev npm wget git
RUN python3 -m pip install --upgrade pip
RUN pip3 install datacube geojson geopandas folium ipyleaflet descartes gpytorch fiona rioxarray pycrs gpytorch torch torchvision ipympl ipyleaflet dask pandas shapely pyproj rtree tslearn fastdtw statsmodels h5py pycolonies

WORKDIR /
COPY ./bin/ml_executor /bin

CMD ["ml_executor", "start"]
