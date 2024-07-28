document.addEventListener('DOMContentLoaded', function() {
    document.querySelector('#comment-form form').addEventListener('submit', function(e) {
        e.preventDefault();

        var commentText = document.querySelector('textarea[name="comment_text"]').value;

        if (commentText.trim() === "") {
            alert('Comment text cannot be empty.');
            return;
        }

        var postID = this.action.split('/').pop();

        fetch(`/view_post/${postID}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                'comment_text': commentText
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            var newComment = data.comment;

            var commentHTML = `
                <div class="comment-sec">
                    <div class="line"></div>
                    <div class="comment-box">
                        <h6><b>${newComment.Username}</b></h6>
                        <br>
                        <p>${newComment.comment_text}</p>
                        <br>
                        <br>
                        <div class="comment-option">
            `;

            if (newComment.Username != "undefined") {
                commentHTML += `
                    <form method="POST">
                        <button type="submit" class="clike-button" data-comment-id="${newComment.CommentID}" data-is-liked="${newComment.IsLiked}">
                            <div class="heartLike" id="clike-heart-${newComment.CommentID}"></div>
                        </button>
                    </form>
                    <b>Likes <span id="clike-count-${newComment.CommentID}">${newComment.like_count}</span> </b>

                    <form method="POST">
                        <button type="submit" class="cdislike-button" data-comment-id="${newComment.CommentID}" data-is-disliked="${newComment.IsDisliked}">
                            <div class="heartDislike" id="cdislike-heart-${newComment.CommentID}"></div>
                        </button>
                    </form>
                    <b>Dislikes <span id="cdislike-count-${newComment.CommentID}">${newComment.dislike_count}</span></b>
                `;
            }

            commentHTML += `
                        </div>
                    </div>
                </div>
            `;

            document.getElementById('comments-list').insertAdjacentHTML('beforeend', commentHTML);

            document.querySelector('textarea[name="comment_text"]').value = '';

            window.location.reload(false); // Refresh the page without scrolling to the top
        })
        .catch(error => {
            console.error('Error:', error);
        });
    });
});
