use crate::create::create_image;

use poise::futures_util::Stream;
use poise::futures_util::StreamExt;
use poise::serenity_prelude::futures;
use serenity::builder::{CreateAttachment, CreateMessage};
use crate::{Error, PoiseContext};
use crate::role::{process_color, random_color};
use crate::config::CONFIG;

async fn autocomplete_name<'a>(_ctx: PoiseContext<'_>, partial: &'a str, ) -> impl Stream<Item = String> + 'a {
    futures::stream::iter(&CONFIG.colors)
        .filter(move |(name, _)| futures::future::ready(name.starts_with(partial)))
        .map(|(name, _)| name.to_owned())
}

#[poise::command(prefix_command, slash_command)]
pub async fn color(
    ctx: PoiseContext<'_>,
    #[description = "The color you want"]
    #[autocomplete = "autocomplete_name"]
    color_choice: Option<String>, ) -> Result<(), Error> {

    match color_choice {
        Some(choice) => {
            if CONFIG.colors.contains_key(choice.as_str()) {
                process_color(ctx, choice.to_string()).await?;
                ctx.say(format!("Assigned color {choice}")).await?;
            } else {
                ctx.say("Unknown color ...").await?;
            }
        },
        None => {
            random_color(&ctx.serenity_context(), &ctx.author_member().await.unwrap()).await?;
            ctx.say(format!("Assigned random color")).await?;
        }
    }
    Ok(())
}

#[poise::command(prefix_command, slash_command)]
pub async fn preview(
    ctx: PoiseContext<'_>,
    #[description = "The preview you want to see"]
    #[autocomplete = "autocomplete_name"]
    choice: String, ) -> Result<(), Error> {

    if CONFIG.colors.contains_key(&choice) {
        let color = CONFIG.colors[&choice];
        let paths = [
            CreateAttachment::bytes(create_image(&color), format!("{choice}.png"))
        ];

        ctx.channel_id().send_files(ctx.http(), paths, CreateMessage::new().content("")).await?;
        ctx.say("Preview send").await?;
    } else {
        ctx.say("Unknown color ...").await?;
    }

    Ok(())
}

#[poise::command(prefix_command, slash_command)]
pub async fn help(ctx: PoiseContext<'_>) -> Result<(), Error> {
    ctx.say("Help for Color-Bot
            `<<color`   [Assign a random color]
            `<<color ColorName`   [Assign the specified color]
            `<<preview ColorName`   [Send a preview image]").await?;

    Ok(())
}