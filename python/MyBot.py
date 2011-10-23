#!/usr/bin/env python
from ants import *

RICH_NONE = -1
RICH_BAD = 0
RICH_UNKNOWN = -10000
RICH_MAX = 1 << 30
RICH_FOOD = - 2000000
RICH_ENEMY = - 1000000
RICH_ENEMY_HILL = - 3000000
RICH_MY_HILL = -1

MAX_ITER = 100

# define a class with a do_turn method
# the Ants.run method will parse and update bot input
# it will also run the do_turn method for us
class MyBot:
    
    def __init__(self):
        # define class level variables, will be remembered between turns
        pass
    
    # do_setup is run once at the start of the game
    # after the bot has received the game settings
    # the ants class is created and setup by the Ants.run method
    def do_setup(self, ants):
        # initialize data structures after learning the game settings
        self.rows = ants.rows
        self.cols = ants.cols
        self.richmap = [[RICH_BAD for col in range(ants.cols)]
                        for row in range(ants.rows)]
        m = ants.map
        r = self.richmap
        for row in range(ants.rows):
            for col in range(ants.cols):
                cur = m[row][col]
                if cur == WATER:
                    r[row][col] = RICH_NONE
        

    def fill_richmap(self, ants):
        m = ants.map
        r = self.richmap
        for row in range(ants.rows):
            for col in range(ants.cols):
                cur = m[row][col]
                if cur == FOOD:
                    r[row][col] = RICH_FOOD
                    continue
                if cur == WATER:
                    r[row][col] = RICH_NONE
                    continue

                if r[row][col] != RICH_NONE:
                    r[row][col] = 0

#                # Remove a pin from the value, if any (except RICH_NONE)
#                if r[row][col] < 0 and r[row][col] != RICH_NONE:
#                    r[row][col] = - r[row][col]

                if (row,col) in ants.hill_list:
                    if ants.hill_list[(row,col)] == MY_ANT:
                        r[row][col] = RICH_MY_HILL
                    else:
                        r[row][col] = RICH_ENEMY_HILL
                    
                    continue

                if (row,col) in ants.ant_list:
                    if ants.ant_list[(row,col)] != MY_ANT:
                        r[row][col] = RICH_ENEMY
                    continue

                if not ants.visible((row,col)) and m[row][col] != WATER:
                    r[row][col] = RICH_UNKNOWN

    def north(self, pos):
        return ( (pos[0] + self.rows -1) % self.rows, pos[1] )

    def south(self, pos):
        return ( (pos[0] + 1) % self.rows, pos[1] )

    def west(self, pos):
        return ( pos[0], (pos[1] + self.cols - 1) % self.cols )

    def east(self, pos):
        return ( pos[0], (pos[1] + 1) % self.cols )

    def rvalue(self, pos, f):
        return self.val(f(pos))

    def val(self, pos):
        res = self.richmap[pos[0]][pos[1]]
        if res != RICH_NONE and res < 0:
            return -res
        return res

    def north_val(self, pos):
        return self.rvalue(pos, self.north)

    def south_val(self, pos):
        return self.rvalue(pos, self.south)

    def west_val(self, pos):
        return self.rvalue(pos, self.west)

    def east_val(self, pos):
        return self.rvalue(pos, self.east)


    def iterate_richmap(self, count):
        l = {}
        for row in range(self.rows):
            for col in range(self.cols):
                if self.richmap[row][col] >= 0:
                    l[(row,col)] = True

        # Do iterations
        for it in range(count):
            if len(l) == 0:
                break
            cur = l
            l = {}
            for pos in cur.keys():
                row = pos[0]
                col = pos[1]
                if self.richmap[row][col] < 0:
                    continue
                vals = [self.north_val(pos), self.south_val(pos), self.west_val(pos), self.east_val(pos)]
                vals = [val for val in vals if val >= 0]
                if len(vals) == 0:
                    continue
                minR = min(vals)
                maxR = max(vals)
                newR = (minR + maxR) // 2
                if self.val(pos) != newR:
                    self.richmap[row][col] = newR
                    l[self.north(pos)] = True
                    l[self.south(pos)] = True
                    l[self.west(pos)] =  True
                    l[self.east(pos)] = True
                

    def dump_richmap(self, filename):
        f = open(filename, "w+")
        for row in range(self.rows):
            for col in range(self.cols):
                f.write("%d " % self.richmap[row][col])
            f.write("\n")
        f.close()

    def update_richmap(self, ants):
        self.fill_richmap(ants)
        self.iterate_richmap(MAX_ITER)
        self.dump_richmap("richmap.txt")
                        

    # do turn is run once per turn
    # the ants class has the game state and is updated by the Ants.run method
    # it also has several helper methods to use
    def do_turn(self, ants):
        # loop through all my ants and try to give them orders
        # the ant_loc is an ant location tuple in (row, col) form
        self.update_richmap(ants)
        bad_locs = {}

        for ant_loc in ants.my_ants():
            # try all directions in given order
            directions = ('n','e','s','w')
            best_dir = 'n'
            best_loc = ant_loc
            best_r = self.val(ant_loc)
            for direction in directions:
                # the destination method will wrap around the map properly
                # and give us a new (row, col) tuple
                new_loc = ants.destination(ant_loc, direction)
                # passable returns true if the location is land
                if (ants.passable(new_loc) and not new_loc in bad_locs and not new_loc in ants.ant_list):
                    # an order is the location of a current ant and a direction
                    r = self.val(new_loc)
                    if r > best_r:
                        best_r = r
                        best_dir = direction
                        best_loc = new_loc


            if best_loc in bad_locs or best_loc in ants.ant_list:
                continue

            ants.issue_order((ant_loc, best_dir))
            bad_locs[best_loc] = True

            # check if we still have time left to calculate more orders
            if ants.time_remaining() < 10:
                break
            
if __name__ == '__main__':
    # psyco will speed up python a little, but is not needed
    try:
        import psyco
        psyco.full()
    except ImportError:
        pass
    
    try:
        # if run is passed a class with a do_turn method, it will do the work
        # this is not needed, in which case you will need to write your own
        # parsing function and your own game state class
        Ants.run(MyBot())
    except KeyboardInterrupt:
        print('ctrl-c, leaving ...')
