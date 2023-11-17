import matplotlib.pyplot as plt
import numpy as np
import PIL
import tensorflow as tf
from tensorflow.keras.optimizers import Adam

from tensorflow import keras
from tensorflow.keras import layers
from tensorflow.keras.models import Sequential

import pathlib
import os
import pickle
import tarfile
import json
import base64
from pathlib import Path

from pycolonies import Colonies
from pycolonies import colonies_client
from pycolonies import func_spec

#colonies, colonyid, colony_prvkey, executorid, executor_prvkey = colonies_client()
process_id = os.environ.get('COLONIES_PROCESS_ID')
#process_b64 = os.environ.get('COLONIES_PROCESS')
#process_bytes = base64.b64decode(process_b64)
#process_str = process_bytes.decode('utf-8')
#process = json.loads(process_str)
#learningrate = process["spec"]["kwargs"]["learning-rate"]
learningrate = 0.001

print("learningrate:", learningrate)

tar_path = "/cfs/data/flower_photos.tgz"
data_dir = Path("/tmp/flower_photos/")
data_dir.mkdir(parents=True, exist_ok=True)

with tarfile.open(tar_path, "r:gz") as tar:
    tar.extractall(path=data_dir)
  
data_dir = Path("/tmp/flower_photos/flower_photos/")

image_count = len(list(data_dir.glob('*/*.jpg')))
print(image_count)

batch_size = 32
img_height = 180
img_width = 180

train_ds = tf.keras.utils.image_dataset_from_directory(
  #Path("/tmp/flower_photos/flower_photos/"),
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

num_classes = len(class_names)

data_augmentation = keras.Sequential(
  [
    layers.RandomFlip("horizontal",
                      input_shape=(img_height,
                                  img_width,
                                  3)),
    layers.RandomRotation(0.1),
    layers.RandomZoom(0.1),
  ]
)

model = Sequential([
  data_augmentation,
  layers.Rescaling(1./255),
  layers.Conv2D(16, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Conv2D(32, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Conv2D(64, 3, padding='same', activation='relu'),
  layers.MaxPooling2D(),
  layers.Dropout(0.2),
  layers.Flatten(),
  layers.Dense(128, activation='relu'),
  layers.Dense(num_classes, name="outputs")
])

optimizer = Adam(learning_rate=learningrate)
model.compile(optimizer=optimizer,
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

file_list = os.listdir(path)

print("Saving results to /cfs/results")
#with open("/cfs/results/" + process_id + ".pickle", 'wb') as file_pi:
with open("/cfs/results/history.pickle", 'wb') as file_pi:
    pickle.dump(history.history, file_pi)

final_train_accuracy = history.history['accuracy'][-1]
final_val_accuracy = history.history['val_accuracy'][-1]

print("final_train_accuracy: ", final_train_accuracy)
print("final_val_accuracy: ", final_val_accuracy)

#colonies.set_output(process_id, [final_train_accuracy, final_val_accuracy], executor_prvkey)
