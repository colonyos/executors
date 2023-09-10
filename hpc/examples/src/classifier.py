import matplotlib.pyplot as plt
import numpy as np
import PIL
import tensorflow as tf

from tensorflow import keras
from tensorflow.keras import layers
from tensorflow.keras.models import Sequential

import pathlib
import os
import pickle
import tarfile
from pathlib import Path


#dataset_url = "https://storage.googleapis.com/download.tensorflow.org/example_images/flower_photos.tgz"
#dataset_url = "file:///cfs/data/flower_photos.tgz"
#data_dir = tf.keras.utils.get_file('/cfs/data/flower_photos.tgz', origin=dataset_url, extract=True)
#data_dir = pathlib.Path(data_dir).with_suffix('')

tar_path = "/cfs/data/flower_photos.tgz"
data_dir = Path("/tmp/flower_photos")
data_dir.mkdir(parents=True, exist_ok=True)

with tarfile.open(tar_path, "r:gz") as tar:
    tar.extractall(path=data_dir)

image_count = len(list(data_dir.glob('*/*.jpg')))
print("number of images: ", image_count)

batch_size = 32
img_height = 180
img_width = 180

train_ds = tf.keras.utils.image_dataset_from_directory(
  data_dir,
  validation_split=0.2,
  subset="training",
  seed=123,
  image_size=(img_height, img_width),
  batch_size=batch_size)

val_ds = tf.keras.utils.image_dataset_from_directory(
  data_dir,
  validation_split=0.2,
  subset="validation",
  seed=123,
  image_size=(img_height, img_width),
  batch_size=batch_size)

class_names = train_ds.class_names
print(class_names)

AUTOTUNE = tf.data.AUTOTUNE

train_ds = train_ds.cache().shuffle(1000).prefetch(buffer_size=AUTOTUNE)
val_ds = val_ds.cache().prefetch(buffer_size=AUTOTUNE)

normalization_layer = layers.Rescaling(1./255)

normalized_ds = train_ds.map(lambda x, y: (normalization_layer(x), y))
image_batch, labels_batch = next(iter(normalized_ds))
first_image = image_batch[0]
# Notice the pixel values are now in `[0,1]`.
print(np.min(first_image), np.max(first_image))

num_classes = len(class_names)

model = Sequential([
  layers.Rescaling(1./255, input_shape=(img_height, img_width, 3)),
  layers.Conv2D(16, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Conv2D(32, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Conv2D(64, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Flatten(),
  layers.Dense(128, activation='relu'),
  layers.Dense(num_classes)
])

model.compile(optimizer='adam',
              loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
              metrics=['accuracy'])

model.summary()

epochs=10
history = model.fit(
  train_ds,
  validation_data=val_ds,
  epochs=epochs
)

path = "/cfs/results"
isExist = os.path.exists(path)
if not isExist:
   os.makedirs(path)

print("Saving results to /cfs/results")
with open('/cfs/results/history.pickle', 'wb') as file_pi:
    pickle.dump(history.history, file_pi)

contents = os.listdir(path)

# Print the contents
for item in contents:
    print(item)
