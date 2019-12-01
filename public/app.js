new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        chatContent: '', // A running list of chat messages displayed on the screen
        username: null, // Our username
        joined: false // True if email and username have been filled in
    },

    created: function() {
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');

        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            self.chatContent += '<div class="chip">'
                //+ msg.sender
                +'counselor'
                + '</div>'
                + msg.message + '<br/>'; // Parse emojis
            var element = document.getElementById('chat-messages');
            element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
        });

    },

    methods: {  
        send: function () {
            if (this.newMsg !== '') {
                //send message
                this.ws.send(
                    JSON.stringify({
                        receiver: 'server',
                        sender: 'KHACH',
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                //show my sent message
                this.chatContent += '<div class="chip">'
                    + this.username
                    + '</div>'
                    +  $('<p>').html(this.newMsg).text() // Strip out html
                     + '<br/>';
                var element = document.getElementById('chat-messages');
                element.value=this.chatContent
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
                this.newMsg = ''; // Reset newMsg
            }
        },
 
        join: function () {
            if (!this.username) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            this.username = $('<p>').html(this.username).text();
            this.joined = true;
            //send introduce to server
            this.ws.send(
                JSON.stringify({
                        receiver: 'server',
                        sender: 'KHACH',
                        message: 'INTRO 1 2'
                    }
                )
            );

        },
              
    }
});