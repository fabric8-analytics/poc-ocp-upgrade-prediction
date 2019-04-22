from flask import Flask
from flask import request


app = Flask(__name__)


@app.route('/setcontext')
def setcontext():
    app.g.context = request.args.get("ctx")
    app.logger.info("Set context to: {}".format(app.g.context))
    return ('', 200)


@app.route('/endcontext')
def endcontext():
    context = request.args.get("ctx")
    if context == app.g.context:
        app.g.context = ""
    app.logger.info("Set context to: {}".format(context))
    return ('', 200)


@app.route('/logcode', methods=['POST'])
def log_code():
    log_string = request.json["fn"]
    app.logger.info(log_string)
    return ('', 200)


def gremlinRestFunction(*args, **kwargs):
    pass


if __name__ == "__main__":
    app.run(debug=True, port=5001)