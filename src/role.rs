use crate::{config::CONFIG, Error, PoiseContext};

use rand::{seq::IteratorRandom, thread_rng};
use serenity::{
    all::{Member, Permissions},
    builder::EditRole,
    http::CacheHttp,
    prelude::Context,
};
use std::{collections::HashMap, hash::Hash};

fn rand_hash<K: Eq + Hash, V>(hash: &HashMap<K, V>) -> &K {
    hash.keys().choose(&mut thread_rng()).unwrap()
}

pub async fn random_color(ctx: &Context, member: &Member) -> Result<(), Error> {
    let guild = match member.guild_id.to_guild_cached(&ctx.cache) {
        Some(g) => g.clone(),
        None => return Ok(()),
    };

    let choice = rand_hash(&CONFIG.colors);
    let role_id = match guild.role_by_name(choice) {
        Some(role) => role.id,
        None => {
            guild
            .create_role(&ctx.http, EditRole::new().name(choice).colour(CONFIG.colors[choice]).permissions(Permissions::empty()))
            .await?
            .id
        },
    };

    if let Some(roles) = member.roles(&ctx.cache) {
        for role in roles.iter().filter(|role| CONFIG.colors.contains_key(&role.name)) {
            member.remove_role(ctx.http(), role.id).await?
        }
    };

    Ok(member.add_role(&ctx.http, role_id).await?)
}

pub async fn process_color(ctx: PoiseContext<'_>, choice: &str) -> Result<(), Error> {
    let guild = match ctx.guild() {
        Some(guild) => guild.clone(),
        None => return Ok(eprintln!("Can't find server ...")),
    };

    let role_id = match guild.role_by_name(choice) {
        Some(role) => role.id,
        None => {
            guild
            .create_role(ctx.http(), EditRole::new().name(choice).colour(CONFIG.colors[choice]).permissions(Permissions::empty()))
            .await?
            .id
        },
    };

    let m = ctx.author_member().await.unwrap();
    if let Some(roles) = m.roles(ctx.cache()) {
        for role in roles.iter().filter(|role| CONFIG.colors.contains_key(&role.name)) {
            m.remove_role(ctx.http(), role.id).await?;
        }
    };

    Ok(m.add_role(ctx.http(), role_id).await?)
}
