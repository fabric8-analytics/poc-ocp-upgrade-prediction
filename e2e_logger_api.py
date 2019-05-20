from flask import Flask
from flask import request
from flask import g
import logging


app = Flask(__name__)
logging.basicConfig(
    format='%(asctime)s %(levelname)-8s %(message)s',
    level=logging.WARNING,
    filename='testruntime.log',
    filemode='w',
    datefmt='%Y-%m-%d %H:%M:%S')
request_context = ""



@app.route('/setcontext')
def setcontext():
    global request_context
    request_context= request.args.get("ctx")
    app.logger.warn("Set context to: {}".format(request_context))
    return ('', 200)


@app.route('/endcontext')
def endcontext():
    global request_context
    context = request.args.get("ctx")
    if context == request_context:
        request_context= ""
    app.logger.warn("End of context: {}".format(context))
    return ('', 200)


if __name__ == "__main__":
    app.run(debug=True, port=5001)
