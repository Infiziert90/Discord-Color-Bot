use crate::{
    config::CONFIG,
    create::create_image,
    role::{process_color, random_color},
    Error, PoiseContext,
};

use poise::{
    futures_util::{Stream, StreamExt},
    serenity_prelude::futures,
    CreateReply,
};
use serenity::builder::CreateAttachment;

async fn autocomplete_name<'a>(_ctx: PoiseContext<'_>, partial: &'a str) -> impl Stream<Item = String> + 'a {
    futures::stream::iter(&CONFIG.colors)
        .filter(move |(name, _)| futures::future::ready(name.starts_with(partial)))
        .map(|(name, _)| name.to_owned())
}

#[poise::command(prefix_command, slash_command)]
pub async fn color(
    ctx: PoiseContext<'_>,
    #[description = "The color you want"]
    #[autocomplete = "autocomplete_name"]
    color_choice: Option<String>,
) -> Result<(), Error> {
    match color_choice {
        Some(choice) => {
            if CONFIG.colors.contains_key(&choice) {
                process_color(ctx, &choice).await?;
                ctx.say(format!("Assigned color {choice}")).await?;
            } else {
                ctx.say("Unknown color ...").await?;
            }
        }
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
    choice: String,
) -> Result<(), Error> {
    if let Some(color) = CONFIG.colors.get(&choice) {
        let paths = [CreateAttachment::bytes(create_image(&color), format!("{choice}.png"))];
        ctx.send(CreateReply {
            content: Some("Preview:".into()),
            attachments: paths.to_vec().into(),
            ..Default::default()
        })
        .await?;
    } else {
        ctx.say("Unknown color ...").await?;
    }

    Ok(())
}

#[poise::command(prefix_command, slash_command)]
pub async fn help(ctx: PoiseContext<'_>) -> Result<(), Error> {
    ctx.say(
        "Help for Color-Bot
            `<<color`   [Assign a random color]
            `<<color ColorName`   [Assign the specified color]
            `<<preview ColorName`   [Send a preview image]",
    )
    .await?;

    Ok(())
}
