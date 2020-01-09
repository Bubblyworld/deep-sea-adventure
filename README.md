# Deep Sea Adventure

This is an implementation of the mechanics of the board game "Deep Sea 
Adventure" by Jun & Goro Sasaki. The point it to computationally explore
the space of possible strategies and evaluate them over thousands of games.

## Overview

The board game consists of an "air-tank" with 25 units of air, 6 character
tokens, 48 treasure tokens and two special 6-sided dice. The faces of the dice
are numbered 1-3, with opposite faces having the same number. The treasure
tokens are split up into five different types:

* 16 empty tokens
* 8 tokens labelled with 1 dot
* 8 tokens labelled with 2 dots
* 8 tokens labelled with 3 dots
* 8 tokens labelled with 4 dots

The probabilities of getting a given number between 2-6 when rolling two of
the special dice are as follows:

* 2: 1/9
* 3: 2/9
* 4: 3/9
* 5: 2/9
* 6: 1/9

The point of the game is to dive into the ocean and search for treasure. The
more treasure you pick up, the slower you move, and the faster the communal
air runs out! The player with the most treasure after 3 rounds of diving is
the winner. The actual rules of the game won't be described here, but if 
you're interested I'd suggest you either buy the game (it's fun!) or read the
source code (less fun).

## Results
Once I've done an analysis of a few strategies, I'll write up some results.
