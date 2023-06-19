# Plot Party Day 4

## Prompt: Water
Description from the plot party website:
> June 15 – Water
>
> A broad theme, will you dare mix a wet media with your pen plotter, or create a work of art reminiscent of water?

## Design

This design combined a flow field and a custom shape, which looks like a stylized water droplet. The noise
for the flow field is generated with OpenSimplex noise, which is builtin to Sketchy. Most of the sliders
control the noise parameters, and the code to handle it were pulled from Sketchy's [noise example](https://github.com/aldernero/sketchy/tree/main/examples/noise).

The other sliders control how many droplets to place in a when pressing "add droplets", and the size and shape
of each droplet. The min/max radius refer to the circular part of the droplet, and the min/max ratio refer to
the ratio of the droplet's overall length to its radius.

As you press "add droplets", droplets are randomly sized and placed in the grid. The direction of the droplet
is mapped to the noise field. You can see what the noise field looks like by pressing "show noise". The 
time to render will increase as you add more droplets, and may become noticeably long if you have many
thousands of droplets. You can start over by pressing "reset droplets".

Here is an example of the UI:

![Screenshot_20230618_201435](https://github.com/aldernero/plotparty-june2023/assets/96601789/71a17d28-1225-41b8-99c0-60b24e6b4010)


There are three colors of droplets, which are related to their size. **Pen plotter tip:** when using multiple
colors, it's convenient to use simple colors like CYMK or RGB, even if the plotted result will use something
different. The reason is that then in Inkscape you can use Edit->Find and find by color hex code, which will
be easy to remember if you use simple colors. Once all paths of a certain color are selected, you can then
add them to a separate layer.

## Results

I used 2 shades of blue and one shade of green. Here is an example:

![20230615_102233](https://github.com/aldernero/plotparty-june2023/assets/96601789/6ba747f9-caa2-40a4-a0fd-ee587d071d49)

