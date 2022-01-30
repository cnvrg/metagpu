import tensorflow as tf
tf.get_logger().setLevel('INFO')
gpus = tf.config.list_physical_devices('GPU')
print(gpus)