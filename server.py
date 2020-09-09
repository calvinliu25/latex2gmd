# Server for latex2gmd by Calvin Liu
# Handles all the Latex tokens and outputs to Github Flavored Markdown
# adapted from provided template in amqp_server.py from the provided project template
# adapted from https://pika.readthedocs.io/en/stable/examples/blocking_consume.html

import pika
import json
import sys
import output

# create C++ vectors for collecting all the tokens from the client
dataList = output.stringVector()
toggleMathModeList = output.boolVector()

# Pika related Section

# declares the connection and channel for later
print("Starting Channel")
connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost"))
channel = connection.channel()
channel.queue_declare(queue="rpc_queue")

# request is a list of tokens
def on_request(ch, method, props, body):
    try:
        request = json.loads(body.decode("utf-8"))
    except json.decoder.JSONDecodeError:
        ch.basic_ack(delivery_tag=method.delivery_tag)
        print("Bad request:", body)
        return

    # collect all the tokens into vectors
    for token in request:
        dataList.push_back(token["Data"])
        toggleMathModeList.push_back(token["ToggleMathMode"])

    response = 1
    body = json.dumps(response).encode("utf-8")

    ch.basic_publish(
        exchange="",
        routing_key=props.reply_to,
        properties=pika.BasicProperties(correlation_id=props.correlation_id),
        body=body,
    )
    ch.basic_ack(delivery_tag=method.delivery_tag)

    # closes channel after receiving one message from the client
    print("Closing Channel")
    channel.stop_consuming()


# Main Function Section

channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue="rpc_queue", on_message_callback=on_request)
print("Awaiting RPC requests")

# exits gracefully if interrupted by ctrl + c when run manually and not by latex2gmd.sh
try:
    channel.start_consuming()
except KeyboardInterrupt:
    channel.stop_consuming()
    connection.close()
    sys.exit("Channel Interrupted")

connection.close()

# Processing Section

gmdText = output.parseTokens(dataList, toggleMathModeList)

# defaults output file name if not given by the user
# will only be seen in the case that this is not called from ./latex2gmd.sh
if len(sys.argv) < 2:
    print("No output file name provided, will default to GMDoutput.md")

    with open("GMDoutput.md", "w") as gmdFile:
        gmdFile.write(gmdText)
else:
    with open(sys.argv[1], "w") as gmdFile:
        gmdFile.write(gmdText)
