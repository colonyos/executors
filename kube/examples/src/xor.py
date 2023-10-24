# import tensorflow library
# Since we'll be using functionalities
# of tensorflow V1 Let us import Tensorflow v1
import tensorflow.compat.v1 as tf
tf.disable_v2_behavior()
import numpy as np
 
# Create placeholders for input X and output Y
X = tf.placeholder(dtype=tf.float32, shape=(4, 2))
Y = tf.placeholder(dtype=tf.float32, shape=(4, 1))
 
# Give training input and label
INPUT_XOR = [[0,0],[0,1],[1,0],[1,1]]
OUTPUT_XOR = [[0],[1],[1],[0]]
 
# Give a standard learning rate and the number
# of epochs the model has to train for.
learning_rate = 0.01
epochs = 10000
 
# Create/Initialize a Hidden Layer variable
with tf.variable_scope('hidden'):
   
      # Initialize weights and biases for the
    # hidden layer randomly whose mean=0 and
    # std_dev=1
    h_w = tf.Variable(tf.truncated_normal([2, 2]), name='weights')
    h_b = tf.Variable(tf.truncated_normal([4, 2]), name='biases')
     
    # Pass the matrix multiplied Input and
    # weights added with Bias to the relu
    # activation function
    h = tf.nn.relu(tf.matmul(X, h_w) + h_b)
 
# Create/Initialize an Output Layer variable
with tf.variable_scope('output'):
       
    # Initialize weights and biases for the
    # output layer randomly whose mean=0 and
    # std_dev=1
    o_w = tf.Variable(tf.truncated_normal([2, 1]), name='weights')
    o_b = tf.Variable(tf.truncated_normal([4, 1]), name='biases')
     
    # Pass the matrix multiplied hidden layer
    # Input and weights added with Bias
    # to a sigmoid activation function
    Y_estimation = tf.nn.sigmoid(tf.matmul(h, o_w) + o_b)
 
# Create/Initialize Loss function variable
with tf.variable_scope('cost'):
   
      # Calculate cost by taking the Root Mean
    # Square between the estimated Y value
    # and the actual Y value
    cost = tf.reduce_mean(tf.squared_difference(Y_estimation, Y))
 
# Create/Initialize Training model variable
with tf.variable_scope('train'):
   
      # Train the model with ADAM Optimizer
    # with the previously initialized learning
    # rate and the cost from the previous variable
    train = tf.train.AdamOptimizer(learning_rate).minimize(cost)
 
# Start a Tensorflow Session
with tf.Session() as session:
   
    # initialize the session variables
    session.run(tf.global_variables_initializer())
    print("Training Started")
     
    # log count
    log_count_frac = epochs/10
    for epoch in range(epochs):
       
        # Training the base network
        session.run(train, feed_dict={X: INPUT_XOR, Y:OUTPUT_XOR})
 
        # log training parameters
        # Print cost for every 1000 epochs
        if epoch % log_count_frac == 0:
            cost_results = session.run(cost, feed_dict={X: INPUT_XOR, Y:OUTPUT_XOR})
            print("Cost of Training at epoch {0} is {1}".format(epoch, cost_results))
 
    print("Training Completed !")
    Y_test = session.run(Y_estimation, feed_dict={X:INPUT_XOR})
    print(np.round(Y_test, decimals=1))
